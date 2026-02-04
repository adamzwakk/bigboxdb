package tools

import (
	"os"
	"fmt"
	"image"
	"path/filepath"
	"strings"
	"sort"
	"os/exec"
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/disintegration/imaging"
	"github.com/sunshineplan/imgconv"
	"github.com/gosimple/slug"

	"github.com/adamzwakk/bigboxdb-server/models"
)

const (
	GatefoldDepthOffset = 0.05
	KTX2Compression     = "etc1s"
	KTX2Quality         = 255
	KTX2QualityLow      = 180
	OutputFormat        = "glb"
)

type GameInfo struct {
	Title		string	 `json:"title"`
	Width       float32  `json:"width"`
	Height      float32  `json:"height"`
	Depth       float32  `json:"depth"`
	BoxType     uint      `json:"box_type"`
}

// AtlasResult holds texture atlas packing results
type AtlasResult struct {
	Atlas              image.Image
	Positions          map[string]image.Point
	Sizes              map[string]image.Point
	Dimensions         image.Point
	OriginalDimensions image.Point
}

// GLTFData holds glTF structure and binary data
type GLTFData struct {
	JSON       map[string]interface{}
	BinaryData *BinaryData
}

// BinaryData holds all binary geometry data
type BinaryData struct {
	Positions         []float32
	Normals           []float32
	UVs               []float32
	Indices           []uint16
	GatefoldPositions []float32
	GatefoldNormals   []float32
	GatefoldUVs       []float32
	GatefoldIndices   []uint16
}

func GenerateGLTFBox(gameInfo *GameInfo, texturePaths []string, outputDir string, lowQuality bool) error {
	upsizeRatio := UpsizeRatio
	ktx2Quality := KTX2Quality
	qualitySuffix := ""

	if lowQuality {
		upsizeRatio = UpsizeRatioLow
		ktx2Quality = KTX2QualityLow
		qualitySuffix = "-low"
	}

	if os.Getenv("APP_ENV") != "production" {
		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("Generating %s quality GLB\n", map[bool]string{true: "LOW", false: "HIGH"}[lowQuality])
		fmt.Printf("Upsize ratio: %d, KTX2 quality: %d\n", upsizeRatio, ktx2Quality)
		fmt.Printf("%s\n\n", strings.Repeat("=", 60))
	}

	// Determine box properties
	boxType := gameInfo.BoxType
	var topWidth *float32
	if boxType == models.FindBoxTypeIDByName("Eidos Trapezoid") {
		var tw float32 = 5.75 
		topWidth = &tw
	}
	gatefoldOnBack := boxType == models.FindBoxTypeIDByName("Big Box With Back Gatefold")

	// Sort texture paths
	boxSortedPaths := make([]string, 6)
	var gatefoldRightPath, gatefoldLeftPath string
	boxSideNames := []string{"front", "back", "top", "bottom", "right", "left"}

	for _, path := range texturePaths {
		filenameLower := strings.ToLower(filepath.Base(path))

		if strings.Contains(filenameLower, "gatefold_right") {
			gatefoldRightPath = path
		} else if strings.Contains(filenameLower, "gatefold_left") {
			gatefoldLeftPath = path
		} else {
			for i, side := range boxSideNames {
				if strings.Contains(filenameLower, side) && boxSortedPaths[i] == "" {
					boxSortedPaths[i] = path
					break
				}
			}
		}
	}

	// Handle missing textures with black placeholders
	for i, path := range boxSortedPaths {
		if path == "" {
			boxSortedPaths[i] = createBlackPlaceholder(outputDir, boxSideNames[i], gameInfo, upsizeRatio, qualitySuffix)
		}
	}

	hasGatefold := gatefoldRightPath != "" && gatefoldLeftPath != ""

	// Pack textures into atlas
	imagesToPack := make(map[string]image.Image)
	
	for i, path := range boxSortedPaths {
		img := loadAndResizeImage(path, lowQuality)
		imagesToPack[boxSideNames[i]] = img
	}

	if hasGatefold {
		// Load gatefold images
		gatefoldRightImg := loadAndResizeImage(gatefoldRightPath, lowQuality)
		gatefoldLeftImg := loadAndResizeImage(gatefoldLeftPath, lowQuality)
		
		// Determine which base face the gatefold replaces
		baseIndex := 0 // front
		if gatefoldOnBack {
			baseIndex = 1 // back
		}
		baseFaceName := boxSideNames[baseIndex]
		
		// Store the base face as gatefold_front
		baseFaceImg := imagesToPack[baseFaceName]
		imagesToPack["gatefold_front"] = baseFaceImg
		
		// Replace the base face with gatefold_right, store gatefold_left as gatefold_back
		if gatefoldOnBack {
			imagesToPack[baseFaceName] = gatefoldLeftImg
			imagesToPack["gatefold_back"] = gatefoldRightImg
		} else {
			imagesToPack[baseFaceName] = gatefoldRightImg
			imagesToPack["gatefold_back"] = gatefoldLeftImg
		}
	}

	// Create atlas filename
	atlasFile := randomString(24) + fmt.Sprintf("-atlas%s.ktx2", qualitySuffix)

	gltfFilename := filepath.Join(outputDir, fmt.Sprintf("box%s.%s", qualitySuffix, OutputFormat))
	atlasFilename := filepath.Join(outputDir, atlasFile)

	atlasResult := packTextures(imagesToPack)

	// Save atlas as KTX2
	if !saveAsKTX2(atlasResult.Atlas, atlasFilename, KTX2Compression, ktx2Quality) {
		return fmt.Errorf("Failed to save KTX2 texture atlas")
	}

	// Generate glTF geometry
	gltfData := generateGeometry(gameInfo, atlasResult, hasGatefold, gatefoldOnBack, topWidth)

	// Save GLB
	saveGLB(gltfData, gltfFilename, atlasFilename, atlasFile)

	if os.Getenv("APP_ENV") != "production" {
		fileInfo, _ := os.Stat(gltfFilename)
		fmt.Printf("%s quality GLB saved: %s (%.1f KB)\n",
			map[bool]string{true: "LOW", false: "HIGH"}[lowQuality],
			gltfFilename,
			float32(fileInfo.Size())/1024)
	}

	return nil
}

func createBlackPlaceholder(outputDir string, sideName string, gameInfo *GameInfo, upsizeRatio int, qualitySuffix string) string {
	var width, height int

	switch sideName {
	case "front", "back":
		width = int(gameInfo.Width * float32(upsizeRatio))
		height = int(gameInfo.Height * float32(upsizeRatio))
	case "left", "right":
		width = int(gameInfo.Depth * float32(upsizeRatio))
		height = int(gameInfo.Height * float32(upsizeRatio))
	case "top", "bottom":
		width = int(gameInfo.Width * float32(upsizeRatio))
		height = int(gameInfo.Depth * float32(upsizeRatio))
	default:
		width, height = 512, 512
	}

	blackImg := imaging.New(width, height, image.Black)
	tempPath := filepath.Join(outputDir, fmt.Sprintf("temp_%s_placeholder%s.webp", sideName, qualitySuffix))
	saveAsWebP(blackImg, tempPath)

	return tempPath
}

func loadAndResizeImage(path string, lowQuality bool) image.Image {
	img, err := imaging.Open(path)
	if err != nil {
		fmt.Printf("Error opening image %s: %v\n", path, err)
		return imaging.New(1, 1, image.Black)
	}

	if lowQuality {
		bounds := img.Bounds()
		newWidth := bounds.Dx() / 2
		newHeight := bounds.Dy() / 2
		img = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
	}

	return img
}

func packTextures(images map[string]image.Image) *AtlasResult {
	// Sort images by height
	type imageEntry struct {
		name string
		img  image.Image
	}

	var entries []imageEntry
	for name, img := range images {
		entries = append(entries, imageEntry{name, img})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].img.Bounds().Dy() > entries[j].img.Bounds().Dy()
	})

	positions := make(map[string]image.Point)
	sizes := make(map[string]image.Point)  // ADDED: Initialize sizes map
	atlasWidth := 0
	atlasHeight := 0
	currentX := 0
	currentY := 0
	rowHeight := 0
	maxWidth := entries[0].img.Bounds().Dx() * 2

	if os.Getenv("APP_ENV") != "production" {
		fmt.Printf("Packing %d textures into atlas...\n", len(entries))
	}

	for _, entry := range entries {
		bounds := entry.img.Bounds()
		imgWidth := bounds.Dx()
		imgHeight := bounds.Dy()

		if currentX+imgWidth > maxWidth {
			currentY += rowHeight
			currentX = 0
			rowHeight = 0
		}

		positions[entry.name] = image.Pt(currentX, currentY)
		sizes[entry.name] = image.Pt(imgWidth, imgHeight)  // ADDED: Store texture size
		
		if os.Getenv("APP_ENV") != "production" {
			fmt.Printf("  '%s': pos=(%d,%d) size=(%d,%d)\n", entry.name, currentX, currentY, imgWidth, imgHeight)
		}
		
		currentX += imgWidth
		rowHeight = max(rowHeight, imgHeight)
		atlasWidth = max(atlasWidth, currentX)
		atlasHeight = currentY + rowHeight
	}

	originalWidth := atlasWidth
	originalHeight := atlasHeight

	// Pad to multiples of 4
	atlasWidth = ((atlasWidth + 3) / 4) * 4
	atlasHeight = ((atlasHeight + 3) / 4) * 4
	
	if os.Getenv("APP_ENV") != "production" {
		fmt.Printf("Atlas dimensions: %dx%d (original: %dx%d)\n", atlasWidth, atlasHeight, originalWidth, originalHeight)
	}

	atlas := imaging.New(atlasWidth, atlasHeight, image.Transparent)

	for name, pos := range positions {
		atlas = imaging.Paste(atlas, images[name], pos)
	}

	return &AtlasResult{
		Atlas:              atlas,
		Positions:          positions,
		Sizes:              sizes,  // ADDED: Include sizes in result
		Dimensions:         image.Pt(atlasWidth, atlasHeight),
		OriginalDimensions: image.Pt(originalWidth, originalHeight),
	}
}

func saveAsKTX2(img image.Image, outputPath, compression string, quality int) bool {
	// Create temporary PNG using imgconv
	tmpFile, err := os.CreateTemp("", "*.png")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return false
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Save as PNG using imgconv
	if err := imgconv.Write(tmpFile, img, &imgconv.FormatOption{Format: imgconv.PNG}); err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		tmpFile.Close()
		return false
	}
	tmpFile.Close()

	// Build toktx command
	args := []string{"--t2", "--genmipmap"}

	if compression == "etc1s" {
		args = append(args, "--encode", "etc1s", "--clevel", "1", "--qlevel", fmt.Sprintf("%d", quality))
	}

	args = append(args, outputPath, tmpPath)

	cmd := exec.Command("toktx", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running toktx: %v\n%s\n", err, output)
		return false
	}

	if os.Getenv("APP_ENV") != "production" {
		fileInfo, _ := os.Stat(outputPath)
		fmt.Printf("KTX2 texture saved: %s (%.1f KB)\n", outputPath, float32(fileInfo.Size())/1024)
	}

	return true
}


func generateGeometry(gameInfo *GameInfo, atlas *AtlasResult, hasGatefold, gatefoldOnBack bool, topWidth *float32) *GLTFData {
	w := gameInfo.Width / 2.0
	h := gameInfo.Height / 2.0
	d := gameInfo.Depth / 2.0

	finalTopWidth := gameInfo.Width
	if topWidth != nil {
		finalTopWidth = *topWidth
	}
	topW := finalTopWidth / 2.0
	isTrapezoid := gameInfo.Width != finalTopWidth

	var trapRatio *float32
	if isTrapezoid {
		ratio := topW / w
		trapRatio = &ratio
	}

	// Box vertices (8 vertices for the box)
	boxVerts := [][3]float32{
		{float32(-w), float32(-h), float32(d)},   // 0
		{float32(w), float32(-h), float32(d)},    // 1
		{float32(topW), float32(h), float32(d)},  // 2
		{float32(-topW), float32(h), float32(d)}, // 3
		{float32(-w), float32(-h), float32(-d)},  // 4
		{float32(w), float32(-h), float32(-d)},   // 5
		{float32(topW), float32(h), float32(-d)}, // 6
		{float32(-topW), float32(h), float32(-d)}, // 7
	}

	binaryData := &BinaryData{}
	
	//boxSideNames := []string{"front", "back", "top", "bottom", "right", "left"}
	
	// Helper function to add UV coordinates
	addUVSet := func(name string, trapezoidRatio *float32, invertTrapezoid, flipHorizontal bool, flipVerticalOverride *bool, rotation int) [][2]float32 {
		pos, ok := atlas.Positions[name]
		if !ok {
			// Return default UVs if texture not found
			fmt.Printf("Warning: Texture '%s' not found in atlas, using default UVs\n", name)
			return [][2]float32{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
		}
		
		size, ok := atlas.Sizes[name]
		if !ok {
			// Fallback size if not found
			fmt.Printf("Warning: Size for texture '%s' not found, using fallback\n", name)
			size = image.Pt(100, 100)
		}
		
		atlasW := float32(atlas.Dimensions.X)
		atlasH := float32(atlas.Dimensions.Y)
		
		// Add 0.5 pixel inset to prevent texture bleeding
		inset := float32(0.5)
		u0Orig := (float32(pos.X) + inset) / atlasW
		u1Orig := (float32(pos.X+size.X) - inset) / atlasW
		v0Orig := (float32(pos.Y) + inset) / atlasH
		v1Orig := (float32(pos.Y+size.Y) - inset) / atlasH
				
		var v0, v1 float32
		if flipVerticalOverride != nil {
			if *flipVerticalOverride {
				v0, v1 = v1Orig, v0Orig
			} else {
				v0, v1 = v0Orig, v1Orig
			}
		} else {
			v0, v1 = v1Orig, v0Orig
		}
		
		u0, u1 := u0Orig, u1Orig
		if flipHorizontal {
			u0, u1 = u1Orig, u0Orig
		}
		
		// Handle rotation
		if rotation != 0 {
			corners := [][2]float32{{u0, v0}, {u1, v0}, {u1, v1}, {u0, v1}}
			switch rotation {
			case 90:
				return [][2]float32{corners[1], corners[2], corners[3], corners[0]}
			case -90, 270:
				return [][2]float32{corners[3], corners[0], corners[1], corners[2]}
			case 180:
				return [][2]float32{corners[2], corners[3], corners[0], corners[1]}
			}
		}
		
		// Handle trapezoid UVs
		if trapezoidRatio != nil {
			uCenter := (u0Orig + u1Orig) / 2
			uHalfWidthBottom := float32(abs(float32(u1Orig - u0Orig))) / 2
			uHalfWidthTop := uHalfWidthBottom * float32(*trapezoidRatio)
			
			if invertTrapezoid {
				if flipHorizontal {
					return [][2]float32{
						{uCenter + uHalfWidthTop, v0},
						{uCenter - uHalfWidthTop, v0},
						{uCenter - uHalfWidthBottom, v1},
						{uCenter + uHalfWidthBottom, v1},
					}
				}
				return [][2]float32{
					{uCenter - uHalfWidthTop, v0},
					{uCenter + uHalfWidthTop, v0},
					{uCenter + uHalfWidthBottom, v1},
					{uCenter - uHalfWidthBottom, v1},
				}
			}
			
			if flipHorizontal {
				return [][2]float32{
					{uCenter + uHalfWidthBottom, v0},
					{uCenter - uHalfWidthBottom, v0},
					{uCenter - uHalfWidthTop, v1},
					{uCenter + uHalfWidthTop, v1},
				}
			}
			return [][2]float32{
				{uCenter - uHalfWidthBottom, v0},
				{uCenter + uHalfWidthBottom, v0},
				{uCenter + uHalfWidthTop, v1},
				{uCenter - uHalfWidthTop, v1},
			}
		}
		
		return [][2]float32{{u0, v0}, {u1, v0}, {u1, v1}, {u0, v1}}
	}
	
	// Helper to add a quad
	addQuad := func(verts [][3]float32, uvCoords [][2]float32, normal [3]float32, isGatefold bool) {
		if isGatefold {
			baseIdx := uint16(len(binaryData.GatefoldPositions) / 3)
			for _, v := range verts {
				binaryData.GatefoldPositions = append(binaryData.GatefoldPositions, v[0], v[1], v[2])
				binaryData.GatefoldNormals = append(binaryData.GatefoldNormals, normal[0], normal[1], normal[2])
			}
			for _, uv := range uvCoords {
				binaryData.GatefoldUVs = append(binaryData.GatefoldUVs, uv[0], uv[1])
			}
			binaryData.GatefoldIndices = append(binaryData.GatefoldIndices, baseIdx, baseIdx+1, baseIdx+2)
			binaryData.GatefoldIndices = append(binaryData.GatefoldIndices, baseIdx, baseIdx+2, baseIdx+3)
		} else {
			baseIdx := uint16(len(binaryData.Positions) / 3)
			for _, v := range verts {
				binaryData.Positions = append(binaryData.Positions, v[0], v[1], v[2])
				binaryData.Normals = append(binaryData.Normals, normal[0], normal[1], normal[2])
			}
			for _, uv := range uvCoords {
				binaryData.UVs = append(binaryData.UVs, uv[0], uv[1])
			}
			binaryData.Indices = append(binaryData.Indices, baseIdx, baseIdx+1, baseIdx+2)
			binaryData.Indices = append(binaryData.Indices, baseIdx, baseIdx+2, baseIdx+3)
		}
	}
	
	// Helper to add a triangle
	addTri := func(verts [][3]float32, uvCoords [][2]float32, normal [3]float32, isGatefold bool) {
		if isGatefold {
			baseIdx := uint16(len(binaryData.GatefoldPositions) / 3)
			for _, v := range verts {
				binaryData.GatefoldPositions = append(binaryData.GatefoldPositions, v[0], v[1], v[2])
				binaryData.GatefoldNormals = append(binaryData.GatefoldNormals, normal[0], normal[1], normal[2])
			}
			for _, uv := range uvCoords {
				binaryData.GatefoldUVs = append(binaryData.GatefoldUVs, uv[0], uv[1])
			}
			binaryData.GatefoldIndices = append(binaryData.GatefoldIndices, baseIdx, baseIdx+1, baseIdx+2)
		} else {
			baseIdx := uint16(len(binaryData.Positions) / 3)
			for _, v := range verts {
				binaryData.Positions = append(binaryData.Positions, v[0], v[1], v[2])
				binaryData.Normals = append(binaryData.Normals, normal[0], normal[1], normal[2])
			}
			for _, uv := range uvCoords {
				binaryData.UVs = append(binaryData.UVs, uv[0], uv[1])
			}
			binaryData.Indices = append(binaryData.Indices, baseIdx, baseIdx+1, baseIdx+2)
		}
	}
	
	// Generate box geometry
	
	// Front face
	uvF := addUVSet("front", trapRatio, false, false, nil, 0)
	if gameInfo.BoxType == models.FindBoxTypeIDByName("Big Box With Vertical Gatefold But Horizontal") {
		uvF = addUVSet("front", trapRatio, false, false, nil, 90)
	}
	if isTrapezoid {
		addTri([][3]float32{boxVerts[0], boxVerts[1], boxVerts[2]}, 
			[][2]float32{uvF[0], uvF[1], uvF[2]}, [3]float32{0, 0, 1}, false)
		addTri([][3]float32{boxVerts[0], boxVerts[2], boxVerts[3]}, 
			[][2]float32{uvF[0], uvF[2], uvF[3]}, [3]float32{0, 0, 1}, false)
	} else {
		addQuad([][3]float32{boxVerts[0], boxVerts[1], boxVerts[2], boxVerts[3]}, 
			uvF, [3]float32{0, 0, 1}, false)
	}
	
	// Back face
	uvBk := addUVSet("back", trapRatio, false, false, nil, 0)
	if isTrapezoid {
		addTri([][3]float32{boxVerts[5], boxVerts[4], boxVerts[7]}, 
			[][2]float32{uvBk[0], uvBk[1], uvBk[2]}, [3]float32{0, 0, -1}, false)
		addTri([][3]float32{boxVerts[5], boxVerts[7], boxVerts[6]}, 
			[][2]float32{uvBk[0], uvBk[2], uvBk[3]}, [3]float32{0, 0, -1}, false)
	} else {
		addQuad([][3]float32{boxVerts[5], boxVerts[4], boxVerts[7], boxVerts[6]}, 
			uvBk, [3]float32{0, 0, -1}, false)
	}
	
	// Right face
	uvR := addUVSet("right", nil, false, false, nil, 0)
	if isTrapezoid {
		addTri([][3]float32{boxVerts[1], boxVerts[5], boxVerts[6]}, 
			[][2]float32{uvR[0], uvR[1], uvR[2]}, [3]float32{1, 0, 0}, false)
		addTri([][3]float32{boxVerts[1], boxVerts[6], boxVerts[2]}, 
			[][2]float32{uvR[0], uvR[2], uvR[3]}, [3]float32{1, 0, 0}, false)
	} else {
		addQuad([][3]float32{boxVerts[1], boxVerts[5], boxVerts[6], boxVerts[2]}, 
			uvR, [3]float32{1, 0, 0}, false)
	}
	
	// Left face
	uvL := addUVSet("left", nil, false, false, nil, 0)
	if isTrapezoid {
		addTri([][3]float32{boxVerts[4], boxVerts[0], boxVerts[3]}, 
			[][2]float32{uvL[0], uvL[1], uvL[2]}, [3]float32{-1, 0, 0}, false)
		addTri([][3]float32{boxVerts[4], boxVerts[3], boxVerts[7]}, 
			[][2]float32{uvL[0], uvL[2], uvL[3]}, [3]float32{-1, 0, 0}, false)
	} else {
		addQuad([][3]float32{boxVerts[4], boxVerts[0], boxVerts[3], boxVerts[7]}, 
			uvL, [3]float32{-1, 0, 0}, false)
	}
	
	// Top face
	uvT := addUVSet("top", nil, false, false, nil, 0)
	addQuad([][3]float32{boxVerts[3], boxVerts[2], boxVerts[6], boxVerts[7]}, 
		uvT, [3]float32{0, 1, 0}, false)
	
	// Bottom face
	uvB := addUVSet("bottom", nil, false, false, nil, 0)
	addQuad([][3]float32{boxVerts[4], boxVerts[5], boxVerts[1], boxVerts[0]}, 
		uvB, [3]float32{0, -1, 0}, false)
	
	// Gatefold geometry
	if hasGatefold {
		gfD := d * GatefoldDepthOffset
		
		var gfVerts [][3]float32
		if gatefoldOnBack {
			gfZ := -d
			gfOffset := -gfD
			gfVerts = [][3]float32{
				{float32(w), float32(-h), float32(gfZ + gfOffset)},
				{float32(-w), float32(-h), float32(gfZ + gfOffset)},
				{float32(-topW), float32(h), float32(gfZ + gfOffset)},
				{float32(topW), float32(h), float32(gfZ + gfOffset)},
				{float32(w), float32(-h), float32(gfZ)},
				{float32(-w), float32(-h), float32(gfZ)},
				{float32(-topW), float32(h), float32(gfZ)},
				{float32(topW), float32(h), float32(gfZ)},
			}
		} else {
			gfZ := d
			gfOffset := gfD
			gfVerts = [][3]float32{
				{float32(-w), float32(-h), float32(gfZ + gfOffset)},
				{float32(w), float32(-h), float32(gfZ + gfOffset)},
				{float32(topW), float32(h), float32(gfZ + gfOffset)},
				{float32(-topW), float32(h), float32(gfZ + gfOffset)},
				{float32(-w), float32(-h), float32(gfZ)},
				{float32(w), float32(-h), float32(gfZ)},
				{float32(topW), float32(h), float32(gfZ)},
				{float32(-topW), float32(h), float32(gfZ)},
			}
		}
		
		// Gatefold front
		uvGf := addUVSet("gatefold_front", trapRatio, false, false, nil, 0)
		if isTrapezoid {
			addTri([][3]float32{gfVerts[0], gfVerts[1], gfVerts[2]}, 
				[][2]float32{uvGf[0], uvGf[1], uvGf[2]}, [3]float32{0, 0, 1}, true)
			addTri([][3]float32{gfVerts[0], gfVerts[2], gfVerts[3]}, 
				[][2]float32{uvGf[0], uvGf[2], uvGf[3]}, [3]float32{0, 0, 1}, true)
		} else {
			addQuad([][3]float32{gfVerts[0], gfVerts[1], gfVerts[2], gfVerts[3]}, 
				uvGf, [3]float32{0, 0, 1}, true)
		}
		
		// Gatefold back
		rotationAngle := 0
		flipGatefoldBack := false
		var flipVert *bool
		
		if !gatefoldOnBack && (gameInfo.BoxType == models.FindBoxTypeIDByName("Eidos Trapezoid") || gameInfo.BoxType == models.FindBoxTypeIDByName("Small Box With Vertical Gatefold")) {
			flipGatefoldBack = true
			f := false
			flipVert = &f
		} else if gameInfo.BoxType == models.FindBoxTypeIDByName("Big Box With Vertical Gatefold But Horizontal") {
			rotationAngle = -90
		} else {
			flipGatefoldBack = isTrapezoid && !gatefoldOnBack
		}
		
		uvGb := addUVSet("gatefold_back", trapRatio, false, flipGatefoldBack, flipVert, rotationAngle)
		if isTrapezoid {
			addTri([][3]float32{gfVerts[5], gfVerts[4], gfVerts[7]}, 
				[][2]float32{uvGb[0], uvGb[1], uvGb[2]}, [3]float32{0, 0, -1}, true)
			addTri([][3]float32{gfVerts[5], gfVerts[7], gfVerts[6]}, 
				[][2]float32{uvGb[0], uvGb[2], uvGb[3]}, [3]float32{0, 0, -1}, true)
		} else {
			addQuad([][3]float32{gfVerts[5], gfVerts[4], gfVerts[7], gfVerts[6]}, 
				uvGb, [3]float32{0, 0, -1}, true)
		}
		
		// Gatefold top (reuse main box top texture)
		uvGt := addUVSet("top", nil, false, false, nil, 0)
		if isTrapezoid {
			addTri([][3]float32{gfVerts[3], gfVerts[2], gfVerts[6]}, 
				[][2]float32{uvGt[0], uvGt[1], uvGt[2]}, [3]float32{0, 1, 0}, true)
			addTri([][3]float32{gfVerts[3], gfVerts[6], gfVerts[7]}, 
				[][2]float32{uvGt[0], uvGt[2], uvGt[3]}, [3]float32{0, 1, 0}, true)
		} else {
			addQuad([][3]float32{gfVerts[3], gfVerts[2], gfVerts[6], gfVerts[7]}, 
				uvGt, [3]float32{0, 1, 0}, true)
		}
		
		// Gatefold bottom (reuse main box bottom texture)
		uvGbot := addUVSet("bottom", nil, false, false, nil, 0)
		if isTrapezoid {
			addTri([][3]float32{gfVerts[1], gfVerts[5], gfVerts[4]}, 
				[][2]float32{uvGbot[2], uvGbot[3], uvGbot[0]}, [3]float32{0, -1, 0}, true)
			addTri([][3]float32{gfVerts[1], gfVerts[4], gfVerts[0]}, 
				[][2]float32{uvGbot[2], uvGbot[0], uvGbot[1]}, [3]float32{0, -1, 0}, true)
		} else {
			addQuad([][3]float32{gfVerts[1], gfVerts[5], gfVerts[4], gfVerts[0]}, 
				uvGbot, [3]float32{0, -1, 0}, true)
		}
		
		// Gatefold right (reuse main box right texture)
		uvGr := addUVSet("right", nil, false, false, nil, 0)
		if isTrapezoid {
			addTri([][3]float32{gfVerts[1], gfVerts[5], gfVerts[6]}, 
				[][2]float32{uvGr[0], uvGr[1], uvGr[2]}, [3]float32{1, 0, 0}, true)
			addTri([][3]float32{gfVerts[1], gfVerts[6], gfVerts[2]}, 
				[][2]float32{uvGr[0], uvGr[2], uvGr[3]}, [3]float32{1, 0, 0}, true)
		} else {
			addQuad([][3]float32{gfVerts[1], gfVerts[5], gfVerts[6], gfVerts[2]}, 
				uvGr, [3]float32{1, 0, 0}, true)
		}
		
		// Gatefold left (reuse main box left texture)
		uvGl := addUVSet("left", nil, false, false, nil, 0)
		if isTrapezoid {
			addTri([][3]float32{gfVerts[4], gfVerts[0], gfVerts[3]}, 
				[][2]float32{uvGl[0], uvGl[1], uvGl[2]}, [3]float32{-1, 0, 0}, true)
			addTri([][3]float32{gfVerts[4], gfVerts[3], gfVerts[7]}, 
				[][2]float32{uvGl[0], uvGl[2], uvGl[3]}, [3]float32{-1, 0, 0}, true)
		} else {
			addQuad([][3]float32{gfVerts[4], gfVerts[0], gfVerts[3], gfVerts[7]}, 
				uvGl, [3]float32{-1, 0, 0}, true)
		}
	}
	
	return &GLTFData{
		JSON:       createGLTFStructure(gameInfo, atlas, hasGatefold, binaryData),
		BinaryData: binaryData,
	}
}

func createGLTFStructure(gameInfo *GameInfo, atlas *AtlasResult, hasGatefold bool, binaryData *BinaryData) map[string]interface{} {
	boxName := slug.Make(gameInfo.Title)
	
	// Calculate bounds for box positions
	minPos := []float32{binaryData.Positions[0], binaryData.Positions[1], binaryData.Positions[2]}
	maxPos := []float32{binaryData.Positions[0], binaryData.Positions[1], binaryData.Positions[2]}
	
	for i := 0; i < len(binaryData.Positions); i += 3 {
		x, y, z := binaryData.Positions[i], binaryData.Positions[i+1], binaryData.Positions[i+2]
		if x < minPos[0] { minPos[0] = x }
		if y < minPos[1] { minPos[1] = y }
		if z < minPos[2] { minPos[2] = z }
		if x > maxPos[0] { maxPos[0] = x }
		if y > maxPos[1] { maxPos[1] = y }
		if z > maxPos[2] { maxPos[2] = z }
	}
	
	// Build nodes - start with box
	nodes := []map[string]interface{}{
		{
			"mesh": 0,
			"name": "Box",
		},
	}
	
	// Build meshes - start with box
	meshes := []map[string]interface{}{
		{
			"name": "Box",
			"primitives": []map[string]interface{}{
				{
					"attributes": map[string]interface{}{
						"POSITION":   0,
						"NORMAL":     1,
						"TEXCOORD_0": 2,
					},
					"indices":  3,
					"material": 0,
				},
			},
		},
	}
	
	// Build accessors - start with box accessors
	accessors := []map[string]interface{}{
		{ // 0: Box POSITION
			"bufferView":    0,
			"componentType": 5126, // FLOAT
			"count":         len(binaryData.Positions) / 3,
			"type":          "VEC3",
			"min":           []float32{minPos[0], minPos[1], minPos[2]},
			"max":           []float32{maxPos[0], maxPos[1], maxPos[2]},
		},
		{ // 1: Box NORMAL
			"bufferView":    1,
			"componentType": 5126, // FLOAT
			"count":         len(binaryData.Normals) / 3,
			"type":          "VEC3",
		},
		{ // 2: Box TEXCOORD_0
			"bufferView":    2,
			"componentType": 5126, // FLOAT
			"count":         len(binaryData.UVs) / 2,
			"type":          "VEC2",
		},
		{ // 3: Box INDICES
			"bufferView":    3,
			"componentType": 5123, // UNSIGNED_SHORT
			"count":         len(binaryData.Indices),
			"type":          "SCALAR",
		},
	}
	
	// Build buffer views
	currentOffset := 0
	bufferViews := []map[string]interface{}{
		{ // 0: Box Positions
			"buffer":     0,
			"byteOffset": currentOffset,
			"byteLength": len(binaryData.Positions) * 4,
			"target":     34962, // ARRAY_BUFFER
		},
	}
	currentOffset += len(binaryData.Positions) * 4
	
	bufferViews = append(bufferViews, map[string]interface{}{ // 1: Box Normals
		"buffer":     0,
		"byteOffset": currentOffset,
		"byteLength": len(binaryData.Normals) * 4,
		"target":     34962,
	})
	currentOffset += len(binaryData.Normals) * 4
	
	bufferViews = append(bufferViews, map[string]interface{}{ // 2: Box UVs
		"buffer":     0,
		"byteOffset": currentOffset,
		"byteLength": len(binaryData.UVs) * 4,
		"target":     34962,
	})
	currentOffset += len(binaryData.UVs) * 4
	
	bufferViews = append(bufferViews, map[string]interface{}{ // 3: Box Indices
		"buffer":     0,
		"byteOffset": currentOffset,
		"byteLength": len(binaryData.Indices) * 2,
		"target":     34963, // ELEMENT_ARRAY_BUFFER
	})
	currentOffset += len(binaryData.Indices) * 2
	
	// Add gatefold if present
	sceneNodes := []int{0}
	if hasGatefold && len(binaryData.GatefoldPositions) > 0 {
		// Calculate gatefold bounds
		gfMinPos := []float32{binaryData.GatefoldPositions[0], binaryData.GatefoldPositions[1], binaryData.GatefoldPositions[2]}
		gfMaxPos := []float32{binaryData.GatefoldPositions[0], binaryData.GatefoldPositions[1], binaryData.GatefoldPositions[2]}
		
		for i := 0; i < len(binaryData.GatefoldPositions); i += 3 {
			x, y, z := binaryData.GatefoldPositions[i], binaryData.GatefoldPositions[i+1], binaryData.GatefoldPositions[i+2]
			if x < gfMinPos[0] { gfMinPos[0] = x }
			if y < gfMinPos[1] { gfMinPos[1] = y }
			if z < gfMinPos[2] { gfMinPos[2] = z }
			if x > gfMaxPos[0] { gfMaxPos[0] = x }
			if y > gfMaxPos[1] { gfMaxPos[1] = y }
			if z > gfMaxPos[2] { gfMaxPos[2] = z }
		}
		
		// Add gatefold node
		nodes = append(nodes, map[string]interface{}{
			"mesh": 1,
			"name": "Gatefold",
		})
		sceneNodes = append(sceneNodes, 1)
		
		// Add gatefold mesh
		meshes = append(meshes, map[string]interface{}{
			"name": "Gatefold",
			"primitives": []map[string]interface{}{
				{
					"attributes": map[string]interface{}{
						"POSITION":   4,
						"NORMAL":     5,
						"TEXCOORD_0": 6,
					},
					"indices":  7,
					"material": 0,
				},
			},
		})
		
		// Add gatefold accessors
		accessors = append(accessors,
			map[string]interface{}{ // 4: Gatefold POSITION
				"bufferView":    4,
				"componentType": 5126,
				"count":         len(binaryData.GatefoldPositions) / 3,
				"type":          "VEC3",
				"min":           []float32{gfMinPos[0], gfMinPos[1], gfMinPos[2]},
				"max":           []float32{gfMaxPos[0], gfMaxPos[1], gfMaxPos[2]},
			},
			map[string]interface{}{ // 5: Gatefold NORMAL
				"bufferView":    5,
				"componentType": 5126,
				"count":         len(binaryData.GatefoldNormals) / 3,
				"type":          "VEC3",
			},
			map[string]interface{}{ // 6: Gatefold TEXCOORD_0
				"bufferView":    6,
				"componentType": 5126,
				"count":         len(binaryData.GatefoldUVs) / 2,
				"type":          "VEC2",
			},
			map[string]interface{}{ // 7: Gatefold INDICES
				"bufferView":    7,
				"componentType": 5123,
				"count":         len(binaryData.GatefoldIndices),
				"type":          "SCALAR",
			},
		)
		
		// Add gatefold buffer views
		bufferViews = append(bufferViews,
			map[string]interface{}{ // 4: Gatefold Positions
				"buffer":     0,
				"byteOffset": currentOffset,
				"byteLength": len(binaryData.GatefoldPositions) * 4,
				"target":     34962,
			},
		)
		currentOffset += len(binaryData.GatefoldPositions) * 4
		
		bufferViews = append(bufferViews,
			map[string]interface{}{ // 5: Gatefold Normals
				"buffer":     0,
				"byteOffset": currentOffset,
				"byteLength": len(binaryData.GatefoldNormals) * 4,
				"target":     34962,
			},
		)
		currentOffset += len(binaryData.GatefoldNormals) * 4
		
		bufferViews = append(bufferViews,
			map[string]interface{}{ // 6: Gatefold UVs
				"buffer":     0,
				"byteOffset": currentOffset,
				"byteLength": len(binaryData.GatefoldUVs) * 4,
				"target":     34962,
			},
		)
		currentOffset += len(binaryData.GatefoldUVs) * 4
		
		bufferViews = append(bufferViews,
			map[string]interface{}{ // 7: Gatefold Indices
				"buffer":     0,
				"byteOffset": currentOffset,
				"byteLength": len(binaryData.GatefoldIndices) * 2,
				"target":     34963,
			},
		)
		currentOffset += len(binaryData.GatefoldIndices) * 2
	}
	
	// Create complete glTF structure
	gltf := map[string]interface{}{
		"asset": map[string]interface{}{
			"version":   "2.0",
			"generator": "BigBoxDB glTF Generator",
		},
		"scene": 0,
		"scenes": []map[string]interface{}{
			{
				"nodes": sceneNodes,
			},
		},
		"nodes": nodes,
		"meshes": meshes,
		"materials": []map[string]interface{}{
			{
				"name": boxName + "-material",
				"pbrMetallicRoughness": map[string]interface{}{
					"baseColorTexture": map[string]interface{}{
						"index": 0,
					},
					"metallicFactor":  0.0,
					"roughnessFactor": 1.0,
				},
			},
		},
		"textures": []map[string]interface{}{
			{
				"source": 0,
			},
		},
		"images": []map[string]interface{}{
			{
				"mimeType": "image/ktx2",
				"bufferView": len(bufferViews), // Will be the next buffer view
			},
		},
		"accessors":   accessors,
		"bufferViews": bufferViews,
		"buffers": []map[string]interface{}{
			{
				"byteLength": currentOffset, // Will be updated with texture data
			},
		},
	}
	
	// Add KTX2 extension
	gltf["extensionsUsed"] = []string{"KHR_texture_basisu"}
	gltf["textures"].([]map[string]interface{})[0]["extensions"] = map[string]interface{}{
		"KHR_texture_basisu": map[string]interface{}{
			"source": 0,
		},
	}
	
	return gltf
}

func saveGLB(gltfData *GLTFData, outputPath, texturePath, textureFile string) error {
	// Pack binary data in correct order matching buffer views
	var buffer bytes.Buffer
	
	// Write box positions
	for _, v := range gltfData.BinaryData.Positions {
		binary.Write(&buffer, binary.LittleEndian, v)
	}
	
	// Write box normals
	for _, v := range gltfData.BinaryData.Normals {
		binary.Write(&buffer, binary.LittleEndian, v)
	}
	
	// Write box UVs
	for _, v := range gltfData.BinaryData.UVs {
		binary.Write(&buffer, binary.LittleEndian, v)
	}
	
	// Write box indices
	for _, v := range gltfData.BinaryData.Indices {
		binary.Write(&buffer, binary.LittleEndian, v)
	}
	
	// Write gatefold data if present
	if len(gltfData.BinaryData.GatefoldPositions) > 0 {
		for _, v := range gltfData.BinaryData.GatefoldPositions {
			binary.Write(&buffer, binary.LittleEndian, v)
		}
		for _, v := range gltfData.BinaryData.GatefoldNormals {
			binary.Write(&buffer, binary.LittleEndian, v)
		}
		for _, v := range gltfData.BinaryData.GatefoldUVs {
			binary.Write(&buffer, binary.LittleEndian, v)
		}
		for _, v := range gltfData.BinaryData.GatefoldIndices {
			binary.Write(&buffer, binary.LittleEndian, v)
		}
	}
	
	// Read texture data
	textureData, err := os.ReadFile(texturePath)
	if err != nil {
		return fmt.Errorf("failed to read texture: %w", err)
	}
	
	// Add buffer view for texture in the JSON
	textureBufferView := map[string]interface{}{
		"buffer":     0,
		"byteOffset": buffer.Len(),
		"byteLength": len(textureData),
	}
	
	// Update the glTF JSON to include texture buffer view
	bufferViews := gltfData.JSON["bufferViews"].([]map[string]interface{})
	bufferViews = append(bufferViews, textureBufferView)
	gltfData.JSON["bufferViews"] = bufferViews
	
	// Update image to reference the buffer view
	images := gltfData.JSON["images"].([]map[string]interface{})
	images[0]["bufferView"] = len(bufferViews) - 1
	gltfData.JSON["images"] = images
	
	// Write texture data to buffer
	buffer.Write(textureData)
	
	// Update total buffer length in JSON
	buffers := gltfData.JSON["buffers"].([]map[string]interface{})
	buffers[0]["byteLength"] = buffer.Len()
	gltfData.JSON["buffers"] = buffers
	
	// Pad buffer to 4-byte alignment
	for buffer.Len()%4 != 0 {
		buffer.WriteByte(0)
	}
	
	// Create JSON chunk
	jsonBytes, err := json.Marshal(gltfData.JSON)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	// Pad JSON to 4-byte alignment with spaces
	for len(jsonBytes)%4 != 0 {
		jsonBytes = append(jsonBytes, ' ')
	}
	
	// Write GLB file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()
	
	// GLB Header (12 bytes)
	// Magic: "glTF" (4 bytes)
	if _, err := f.Write([]byte("glTF")); err != nil {
		return err
	}
	
	// Version: 2 (4 bytes)
	if err := binary.Write(f, binary.LittleEndian, uint32(2)); err != nil {
		return err
	}
	
	// Total length (4 bytes)
	totalLength := 12 + 8 + len(jsonBytes) + 8 + buffer.Len()
	if err := binary.Write(f, binary.LittleEndian, uint32(totalLength)); err != nil {
		return err
	}
	
	// JSON Chunk (8 bytes header + data)
	// Chunk length
	if err := binary.Write(f, binary.LittleEndian, uint32(len(jsonBytes))); err != nil {
		return err
	}
	
	// Chunk type: "JSON"
	if _, err := f.Write([]byte("JSON")); err != nil {
		return err
	}
	
	// Chunk data
	if _, err := f.Write(jsonBytes); err != nil {
		return err
	}
	
	// Binary Chunk (8 bytes header + data)
	// Chunk length
	if err := binary.Write(f, binary.LittleEndian, uint32(buffer.Len())); err != nil {
		return err
	}
	
	// Chunk type: "BIN\x00"
	if _, err := f.Write([]byte("BIN\x00")); err != nil {
		return err
	}
	
	// Chunk data
	if _, err := f.Write(buffer.Bytes()); err != nil {
		return err
	}
	
	return nil
}

func CleanupKTX2Files(webdir string) {
	filepath.Walk(webdir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".ktx2") {
			os.Remove(path)
		}
		return nil
	})
}