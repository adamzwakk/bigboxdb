package tools

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gosimple/slug"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"github.com/sunshineplan/imgconv"

	"github.com/adamzwakk/bigboxdb-server/models"
)

const (
	GatefoldDepthOffset = 0.05
	KTX2Compression     = "etc1s"
	KTX2Quality         = 255
	KTX2QualityLow      = 180
	OutputFormat        = "glb"
)

// GatefoldMode describes how gatefolds are arranged on a box
type GatefoldMode int

const (
	GatefoldNone         GatefoldMode = iota
	GatefoldSingleFront               // One full-width flap on front
	GatefoldSingleBack                // One full-width flap on back
	GatefoldDoubleFront               // Two half-width flaps on front (book/double-door)
	GatefoldFrontAndBack              // One full-width flap on front AND one on back
)

type GameInfo struct {
	Title   string  `json:"title"`
	Width   float32 `json:"width"`
	Height  float32 `json:"height"`
	Depth   float32 `json:"depth"`
	BoxType uint    `json:"box_type"`
}

// AtlasResult holds texture atlas packing results
type AtlasResult struct {
	Atlas              image.Image
	Positions          map[string]image.Point
	Sizes              map[string]image.Point
	Dimensions         image.Point
	OriginalDimensions image.Point
}

// MeshPart holds the geometry data for a single distinct object (e.g. "Box", "GatefoldFront")
// Each part becomes a separately named node in the glTF so Three.js can target it for animation
type MeshPart struct {
	Name      string
	Positions [][3]float32
	Normals   [][3]float32
	UVs       [][2]float32
	Indices   []uint16
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

	// Determine gatefold mode from box type
	gatefoldMode := determineGatefoldMode(boxType)

	// Sort texture paths — detect all gatefold texture variants
	boxSortedPaths := make([]string, 6)
	gatefoldPaths := make(map[string]string) // key: "left", "right", "front_left", "front_right", "back"
	boxSideNames := []string{"front", "back", "top", "bottom", "right", "left"}

	for _, path := range texturePaths {
		filenameLower := strings.ToLower(filepath.Base(path))

		switch {
		case strings.Contains(filenameLower, "gatefold_front_left"):
			gatefoldPaths["front_left"] = path
		case strings.Contains(filenameLower, "gatefold_front_right"):
			gatefoldPaths["front_right"] = path
		case strings.Contains(filenameLower, "gatefold_back"):
			gatefoldPaths["back"] = path
		case strings.Contains(filenameLower, "gatefold_right"):
			gatefoldPaths["right"] = path
		case strings.Contains(filenameLower, "gatefold_left"):
			gatefoldPaths["left"] = path
		default:
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

	// Determine if we actually have gatefold textures for the requested mode
	hasGatefold := false
	switch gatefoldMode {
	case GatefoldSingleFront:
		// Need legacy left+right, or front_left+front_right
		hasGatefold = (gatefoldPaths["left"] != "" && gatefoldPaths["right"] != "") ||
			(gatefoldPaths["front_left"] != "" && gatefoldPaths["front_right"] != "")
	case GatefoldSingleBack:
		hasGatefold = (gatefoldPaths["left"] != "" && gatefoldPaths["right"] != "") ||
			(gatefoldPaths["back"] != "")
	case GatefoldDoubleFront:
		hasGatefold = gatefoldPaths["front_left"] != "" && gatefoldPaths["front_right"] != ""
	case GatefoldFrontAndBack:
		// Need textures for both front and back flaps
		hasFront := (gatefoldPaths["left"] != "" && gatefoldPaths["right"] != "") ||
			(gatefoldPaths["front_left"] != "" && gatefoldPaths["front_right"] != "")
		hasBack := gatefoldPaths["back"] != "" ||
			(gatefoldPaths["left"] != "" && gatefoldPaths["right"] != "")
		hasGatefold = hasFront && hasBack
	}

	if !hasGatefold {
		gatefoldMode = GatefoldNone
	}

	// Pack textures into atlas
	imagesToPack := make(map[string]image.Image)

	for i, path := range boxSortedPaths {
		img := loadAndResizeImage(path, lowQuality)
		imagesToPack[boxSideNames[i]] = img
	}

	if hasGatefold {
		switch gatefoldMode {
		case GatefoldSingleFront:
			// Resolve texture paths — prefer new naming, fall back to legacy
			rightPath := gatefoldPaths["front_right"]
			if rightPath == "" {
				rightPath = gatefoldPaths["right"]
			}
			leftPath := gatefoldPaths["front_left"]
			if leftPath == "" {
				leftPath = gatefoldPaths["left"]
			}

			gatefoldRightImg := loadAndResizeImage(rightPath, lowQuality)
			gatefoldLeftImg := loadAndResizeImage(leftPath, lowQuality)

			// The original front face becomes the inside of the gatefold
			baseFaceImg := imagesToPack["front"]
			imagesToPack["gatefold_front_inner"] = baseFaceImg

			// The outer face shows gatefold_right, the back shows gatefold_left
			imagesToPack["front"] = gatefoldRightImg
			imagesToPack["gatefold_front_back"] = gatefoldLeftImg

		case GatefoldSingleBack:
			rightPath := gatefoldPaths["right"]
			leftPath := gatefoldPaths["left"]

			gatefoldRightImg := loadAndResizeImage(rightPath, lowQuality)
			gatefoldLeftImg := loadAndResizeImage(leftPath, lowQuality)

			// The original back face becomes the inside of the gatefold
			baseFaceImg := imagesToPack["back"]
			imagesToPack["gatefold_back_inner"] = baseFaceImg

			// gatefold_left replaces the back face, gatefold_right is the outer back
			imagesToPack["back"] = gatefoldLeftImg
			imagesToPack["gatefold_back_back"] = gatefoldRightImg

		case GatefoldDoubleFront:
			frontLeftImg := loadAndResizeImage(gatefoldPaths["front_left"], lowQuality)
			frontRightImg := loadAndResizeImage(gatefoldPaths["front_right"], lowQuality)

			// The original front face becomes the inside (visible when doors open)
			baseFaceImg := imagesToPack["front"]
			imagesToPack["gatefold_double_inner"] = baseFaceImg

			// Each flap's outer face
			imagesToPack["gatefold_front_left"] = frontLeftImg
			imagesToPack["gatefold_front_right"] = frontRightImg

			// Optional: if there's a gatefold_back texture, use it for the back of the flaps
			if gatefoldPaths["back"] != "" {
				gatefoldBackImg := loadAndResizeImage(gatefoldPaths["back"], lowQuality)
				imagesToPack["gatefold_double_back"] = gatefoldBackImg
			}

		case GatefoldFrontAndBack:
			// Front flap — prefer new naming, fall back to legacy
			frontRightPath := gatefoldPaths["front_right"]
			if frontRightPath == "" {
				frontRightPath = gatefoldPaths["right"]
			}
			frontLeftPath := gatefoldPaths["front_left"]
			if frontLeftPath == "" {
				frontLeftPath = gatefoldPaths["left"]
			}

			frontRightImg := loadAndResizeImage(frontRightPath, lowQuality)
			frontLeftImg := loadAndResizeImage(frontLeftPath, lowQuality)

			// Original front face → inside of front gatefold
			frontBaseImg := imagesToPack["front"]
			imagesToPack["gatefold_front_inner"] = frontBaseImg
			imagesToPack["front"] = frontRightImg
			imagesToPack["gatefold_front_back"] = frontLeftImg

			// Back flap
			backPath := gatefoldPaths["back"]
			if backPath == "" {
				backPath = gatefoldPaths["right"]
			}
			backImg := loadAndResizeImage(backPath, lowQuality)

			// Original back face → inside of back gatefold
			backBaseImg := imagesToPack["back"]
			imagesToPack["gatefold_back_inner"] = backBaseImg
			imagesToPack["gatefold_back_back"] = backImg
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

	// Generate array of distinct MeshParts
	meshParts := generateGeometry(gameInfo, atlasResult, gatefoldMode, topWidth)

	// Generate glTF structured document
	doc, err := generateGLTFDocument(gameInfo, meshParts, atlasFilename)
	if err != nil {
		return fmt.Errorf("failed to build gltf document: %w", err)
	}

	// Save GLB using qmuntal/gltf
	if err := gltf.SaveBinary(doc, gltfFilename); err != nil {
		return fmt.Errorf("failed to save GLB: %w", err)
	}

	if os.Getenv("APP_ENV") != "production" {
		fileInfo, _ := os.Stat(gltfFilename)
		fmt.Printf("%s quality GLB saved: %s (%.1f KB)\n",
			map[bool]string{true: "LOW", false: "HIGH"}[lowQuality],
			gltfFilename,
			float32(fileInfo.Size())/1024)
	}

	return nil
}

// determineGatefoldMode maps box type names to gatefold configurations
func determineGatefoldMode(boxType uint) GatefoldMode {
	switch boxType {
	case models.FindBoxTypeIDByName("Big Box With Back Gatefold"):
		return GatefoldSingleBack
	case models.FindBoxTypeIDByName("Big Box With Double Gatefold"):
		return GatefoldDoubleFront
	case models.FindBoxTypeIDByName("Big Box With Front And Back Gatefold"):
		return GatefoldFrontAndBack
	default:
		// All other gatefold-capable types use single front
		// The caller will check if gatefold textures actually exist
		return GatefoldSingleFront
	}
}

// generateGLTFDocument dynamically loops through N amount of MeshParts
func generateGLTFDocument(gameInfo *GameInfo, parts []*MeshPart, texturePath string) (*gltf.Document, error) {
	doc := gltf.NewDocument()
	doc.Asset.Generator = "BigBoxDB glTF Generator"
	doc.ExtensionsUsed = append(doc.ExtensionsUsed, "KHR_texture_basisu")

	var sceneNodes []int

	// 1. Generate Buffers, BufferViews, Accessors, Meshes, and Nodes for each piece
	for _, part := range parts {
		if len(part.Positions) == 0 {
			continue
		}

		// Let modeler calculate byte offsets and sizes automatically
		posAccessor := modeler.WritePosition(doc, part.Positions)
		normAccessor := modeler.WriteNormal(doc, part.Normals)
		uvAccessor := modeler.WriteTextureCoord(doc, part.UVs)
		indicesAccessor := modeler.WriteIndices(doc, part.Indices)

		// Create the mesh for this specific part
		doc.Meshes = append(doc.Meshes, &gltf.Mesh{
			Name: part.Name,
			Primitives: []*gltf.Primitive{
				{
					Indices: gltf.Index(indicesAccessor),
					Attributes: gltf.PrimitiveAttributes{
						gltf.POSITION:   posAccessor,
						gltf.NORMAL:     normAccessor,
						gltf.TEXCOORD_0: uvAccessor,
					},
					Material: gltf.Index(0), // Uses the shared atlas material
				},
			},
		})

		meshIndex := len(doc.Meshes) - 1

		// Create a Node that Three.js can target by name
		doc.Nodes = append(doc.Nodes, &gltf.Node{
			Name: part.Name,
			Mesh: gltf.Index(meshIndex),
		})

		nodeIndex := len(doc.Nodes) - 1
		sceneNodes = append(sceneNodes, nodeIndex)
	}

	doc.Scenes = append(doc.Scenes, &gltf.Scene{Nodes: sceneNodes})
	doc.Scene = gltf.Index(0)

	// 2. Embed the KTX2 Texture Data
	textureData, err := os.ReadFile(texturePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read texture for embedding: %w", err)
	}

	if len(doc.Buffers) == 0 {
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
	}

	bufferIndex := 0
	byteOffset := len(doc.Buffers[bufferIndex].Data)

	// Write texture to the buffer and ensure 4-byte alignment
	doc.Buffers[bufferIndex].Data = append(doc.Buffers[bufferIndex].Data, textureData...)
	for len(doc.Buffers[bufferIndex].Data)%4 != 0 {
		doc.Buffers[bufferIndex].Data = append(doc.Buffers[bufferIndex].Data, 0)
	}
	doc.Buffers[bufferIndex].ByteLength = len(doc.Buffers[bufferIndex].Data)

	// Create a BufferView pointing to the texture bytes
	doc.BufferViews = append(doc.BufferViews, &gltf.BufferView{
		Buffer:     bufferIndex,
		ByteOffset: byteOffset,
		ByteLength: len(textureData),
	})
	bvIndex := len(doc.BufferViews) - 1

	// 3. Map the Material tree
	doc.Images = append(doc.Images, &gltf.Image{
		MimeType:   "image/ktx2",
		BufferView: gltf.Index(bvIndex),
	})

	doc.Textures = append(doc.Textures, &gltf.Texture{
		Extensions: gltf.Extensions{
			"KHR_texture_basisu": map[string]interface{}{
				"source": 0,
			},
		},
	})

	doc.Materials = append(doc.Materials, &gltf.Material{
		Name: slug.Make(gameInfo.Title) + "-material",
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorTexture: &gltf.TextureInfo{Index: 0},
			MetallicFactor:   gltf.Float(0.0),
			RoughnessFactor:  gltf.Float(1.0),
		},
	})

	return doc, nil
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
	sizes := make(map[string]image.Point)
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
		sizes[entry.name] = image.Pt(imgWidth, imgHeight)

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
		Sizes:              sizes,
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

func generateGeometry(gameInfo *GameInfo, atlas *AtlasResult, gatefoldMode GatefoldMode, topWidth *float32) []*MeshPart {
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
		{float32(-w), float32(-h), float32(d)},    // 0
		{float32(w), float32(-h), float32(d)},     // 1
		{float32(topW), float32(h), float32(d)},   // 2
		{float32(-topW), float32(h), float32(d)},  // 3
		{float32(-w), float32(-h), float32(-d)},   // 4
		{float32(w), float32(-h), float32(-d)},    // 5
		{float32(topW), float32(h), float32(-d)},  // 6
		{float32(-topW), float32(h), float32(-d)}, // 7
	}

	// Dynamic slice of mesh parts
	parts := []*MeshPart{}

	boxMesh := &MeshPart{Name: "Box"}
	parts = append(parts, boxMesh)

	// Helper function to add UV coordinates
	addUVSet := func(name string, trapezoidRatio *float32, invertTrapezoid, flipHorizontal bool, flipVerticalOverride *bool, rotation int) [][2]float32 {
		pos, ok := atlas.Positions[name]
		if !ok {
			fmt.Printf("Warning: Texture '%s' not found in atlas, using default UVs\n", name)
			return [][2]float32{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
		}

		size, ok := atlas.Sizes[name]
		if !ok {
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
			uHalfWidthBottom := float32(abs(float32(u1Orig-u0Orig))) / 2
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

	// Helper to add a quad specifically targeting a MeshPart
	addQuad := func(mesh *MeshPart, verts [][3]float32, uvCoords [][2]float32, normal [3]float32) {
		baseIdx := uint16(len(mesh.Positions))
		for _, v := range verts {
			mesh.Positions = append(mesh.Positions, v)
			mesh.Normals = append(mesh.Normals, normal)
		}
		for _, uv := range uvCoords {
			mesh.UVs = append(mesh.UVs, uv)
		}
		mesh.Indices = append(mesh.Indices, baseIdx, baseIdx+1, baseIdx+2)
		mesh.Indices = append(mesh.Indices, baseIdx, baseIdx+2, baseIdx+3)
	}

	// Helper to add a triangle specifically targeting a MeshPart
	addTri := func(mesh *MeshPart, verts [][3]float32, uvCoords [][2]float32, normal [3]float32) {
		baseIdx := uint16(len(mesh.Positions))
		for _, v := range verts {
			mesh.Positions = append(mesh.Positions, v)
			mesh.Normals = append(mesh.Normals, normal)
		}
		for _, uv := range uvCoords {
			mesh.UVs = append(mesh.UVs, uv)
		}
		mesh.Indices = append(mesh.Indices, baseIdx, baseIdx+1, baseIdx+2)
	}

	// addGatefoldPanel generates a full 6-face panel (front, back, top, bottom, left, right)
	// for a gatefold flap. frontTex/backTex are the atlas keys for the unique faces;
	// edges reuse the main box textures.
	addGatefoldPanel := func(mesh *MeshPart, verts [][3]float32, frontTex, backTex string,
		frontNormal, backNormal [3]float32, trapRatio *float32,
		flipBack bool, flipVert *bool, backRotation int) {

		// Front face (verts 0-3)
		uvFront := addUVSet(frontTex, trapRatio, false, false, nil, 0)
		if isTrapezoid {
			addTri(mesh, [][3]float32{verts[0], verts[1], verts[2]},
				[][2]float32{uvFront[0], uvFront[1], uvFront[2]}, frontNormal)
			addTri(mesh, [][3]float32{verts[0], verts[2], verts[3]},
				[][2]float32{uvFront[0], uvFront[2], uvFront[3]}, frontNormal)
		} else {
			addQuad(mesh, [][3]float32{verts[0], verts[1], verts[2], verts[3]},
				uvFront, frontNormal)
		}

		// Back face (verts 4-7)
		uvBack := addUVSet(backTex, trapRatio, false, flipBack, flipVert, backRotation)
		if isTrapezoid {
			addTri(mesh, [][3]float32{verts[5], verts[4], verts[7]},
				[][2]float32{uvBack[0], uvBack[1], uvBack[2]}, backNormal)
			addTri(mesh, [][3]float32{verts[5], verts[7], verts[6]},
				[][2]float32{uvBack[0], uvBack[2], uvBack[3]}, backNormal)
		} else {
			addQuad(mesh, [][3]float32{verts[5], verts[4], verts[7], verts[6]},
				uvBack, backNormal)
		}

		// Top face (reuse box top texture)
		uvTop := addUVSet("top", nil, false, false, nil, 0)
		if isTrapezoid {
			addTri(mesh, [][3]float32{verts[3], verts[2], verts[6]},
				[][2]float32{uvTop[0], uvTop[1], uvTop[2]}, [3]float32{0, 1, 0})
			addTri(mesh, [][3]float32{verts[3], verts[6], verts[7]},
				[][2]float32{uvTop[0], uvTop[2], uvTop[3]}, [3]float32{0, 1, 0})
		} else {
			addQuad(mesh, [][3]float32{verts[3], verts[2], verts[6], verts[7]},
				uvTop, [3]float32{0, 1, 0})
		}

		// Bottom face (reuse box bottom texture)
		uvBot := addUVSet("bottom", nil, false, false, nil, 0)
		if isTrapezoid {
			addTri(mesh, [][3]float32{verts[1], verts[5], verts[4]},
				[][2]float32{uvBot[2], uvBot[3], uvBot[0]}, [3]float32{0, -1, 0})
			addTri(mesh, [][3]float32{verts[1], verts[4], verts[0]},
				[][2]float32{uvBot[2], uvBot[0], uvBot[1]}, [3]float32{0, -1, 0})
		} else {
			addQuad(mesh, [][3]float32{verts[1], verts[5], verts[4], verts[0]},
				uvBot, [3]float32{0, -1, 0})
		}

		// Right edge (reuse box right texture)
		uvRight := addUVSet("right", nil, false, false, nil, 0)
		if isTrapezoid {
			addTri(mesh, [][3]float32{verts[1], verts[5], verts[6]},
				[][2]float32{uvRight[0], uvRight[1], uvRight[2]}, [3]float32{1, 0, 0})
			addTri(mesh, [][3]float32{verts[1], verts[6], verts[2]},
				[][2]float32{uvRight[0], uvRight[2], uvRight[3]}, [3]float32{1, 0, 0})
		} else {
			addQuad(mesh, [][3]float32{verts[1], verts[5], verts[6], verts[2]},
				uvRight, [3]float32{1, 0, 0})
		}

		// Left edge (reuse box left texture)
		uvLeft := addUVSet("left", nil, false, false, nil, 0)
		if isTrapezoid {
			addTri(mesh, [][3]float32{verts[4], verts[0], verts[3]},
				[][2]float32{uvLeft[0], uvLeft[1], uvLeft[2]}, [3]float32{-1, 0, 0})
			addTri(mesh, [][3]float32{verts[4], verts[3], verts[7]},
				[][2]float32{uvLeft[0], uvLeft[2], uvLeft[3]}, [3]float32{-1, 0, 0})
		} else {
			addQuad(mesh, [][3]float32{verts[4], verts[0], verts[3], verts[7]},
				uvLeft, [3]float32{-1, 0, 0})
		}
	}

	// Generate Box Geometry
	// Front face
	uvF := addUVSet("front", trapRatio, false, false, nil, 0)
	if gameInfo.BoxType == models.FindBoxTypeIDByName("Big Box With Vertical Gatefold But Horizontal") {
		uvF = addUVSet("front", trapRatio, false, false, nil, 90)
	}
	if isTrapezoid {
		addTri(boxMesh, [][3]float32{boxVerts[0], boxVerts[1], boxVerts[2]},
			[][2]float32{uvF[0], uvF[1], uvF[2]}, [3]float32{0, 0, 1})
		addTri(boxMesh, [][3]float32{boxVerts[0], boxVerts[2], boxVerts[3]},
			[][2]float32{uvF[0], uvF[2], uvF[3]}, [3]float32{0, 0, 1})
	} else {
		addQuad(boxMesh, [][3]float32{boxVerts[0], boxVerts[1], boxVerts[2], boxVerts[3]},
			uvF, [3]float32{0, 0, 1})
	}

	// Back face
	uvBk := addUVSet("back", trapRatio, false, false, nil, 0)
	if isTrapezoid {
		addTri(boxMesh, [][3]float32{boxVerts[5], boxVerts[4], boxVerts[7]},
			[][2]float32{uvBk[0], uvBk[1], uvBk[2]}, [3]float32{0, 0, -1})
		addTri(boxMesh, [][3]float32{boxVerts[5], boxVerts[7], boxVerts[6]},
			[][2]float32{uvBk[0], uvBk[2], uvBk[3]}, [3]float32{0, 0, -1})
	} else {
		addQuad(boxMesh, [][3]float32{boxVerts[5], boxVerts[4], boxVerts[7], boxVerts[6]},
			uvBk, [3]float32{0, 0, -1})
	}

	// Right face
	uvR := addUVSet("right", nil, false, false, nil, 0)
	if isTrapezoid {
		addTri(boxMesh, [][3]float32{boxVerts[1], boxVerts[5], boxVerts[6]},
			[][2]float32{uvR[0], uvR[1], uvR[2]}, [3]float32{1, 0, 0})
		addTri(boxMesh, [][3]float32{boxVerts[1], boxVerts[6], boxVerts[2]},
			[][2]float32{uvR[0], uvR[2], uvR[3]}, [3]float32{1, 0, 0})
	} else {
		addQuad(boxMesh, [][3]float32{boxVerts[1], boxVerts[5], boxVerts[6], boxVerts[2]},
			uvR, [3]float32{1, 0, 0})
	}

	// Left face
	uvL := addUVSet("left", nil, false, false, nil, 0)
	if isTrapezoid {
		addTri(boxMesh, [][3]float32{boxVerts[4], boxVerts[0], boxVerts[3]},
			[][2]float32{uvL[0], uvL[1], uvL[2]}, [3]float32{-1, 0, 0})
		addTri(boxMesh, [][3]float32{boxVerts[4], boxVerts[3], boxVerts[7]},
			[][2]float32{uvL[0], uvL[2], uvL[3]}, [3]float32{-1, 0, 0})
	} else {
		addQuad(boxMesh, [][3]float32{boxVerts[4], boxVerts[0], boxVerts[3], boxVerts[7]},
			uvL, [3]float32{-1, 0, 0})
	}

	// Top face
	uvT := addUVSet("top", nil, false, false, nil, 0)
	addQuad(boxMesh, [][3]float32{boxVerts[3], boxVerts[2], boxVerts[6], boxVerts[7]},
		uvT, [3]float32{0, 1, 0})

	// Bottom face
	uvB := addUVSet("bottom", nil, false, false, nil, 0)
	addQuad(boxMesh, [][3]float32{boxVerts[4], boxVerts[5], boxVerts[1], boxVerts[0]},
		uvB, [3]float32{0, -1, 0})

	// ========================
	// Gatefold geometry
	// ========================
	gfD := d * GatefoldDepthOffset

	switch gatefoldMode {
	case GatefoldSingleFront:
		// Single full-width flap on front — node name: "GatefoldFront"
		gfMesh := &MeshPart{Name: "GatefoldFront"}
		parts = append(parts, gfMesh)

		gfZ := d
		gfOffset := gfD
		gfVerts := [][3]float32{
			{float32(-w), float32(-h), float32(gfZ + gfOffset)},
			{float32(w), float32(-h), float32(gfZ + gfOffset)},
			{float32(topW), float32(h), float32(gfZ + gfOffset)},
			{float32(-topW), float32(h), float32(gfZ + gfOffset)},
			{float32(-w), float32(-h), float32(gfZ)},
			{float32(w), float32(-h), float32(gfZ)},
			{float32(topW), float32(h), float32(gfZ)},
			{float32(-topW), float32(h), float32(gfZ)},
		}

		// Determine back face flip/rotation based on box type
		flipBack := false
		var flipVert *bool
		backRotation := 0

		if gameInfo.BoxType == models.FindBoxTypeIDByName("Eidos Trapezoid") || gameInfo.BoxType == models.FindBoxTypeIDByName("Small Box With Vertical Gatefold") {
			flipBack = true
			f := false
			flipVert = &f
		} else if gameInfo.BoxType == models.FindBoxTypeIDByName("Big Box With Vertical Gatefold But Horizontal") {
			backRotation = -90
		} else {
			flipBack = isTrapezoid
		}

		addGatefoldPanel(gfMesh, gfVerts,
			"gatefold_front_inner", "gatefold_front_back",
			[3]float32{0, 0, 1}, [3]float32{0, 0, -1},
			trapRatio, flipBack, flipVert, backRotation)

	case GatefoldSingleBack:
		// Single full-width flap on back — node name: "GatefoldBack"
		gfMesh := &MeshPart{Name: "GatefoldBack"}
		parts = append(parts, gfMesh)

		gfZ := -d
		gfOffset := -gfD
		gfVerts := [][3]float32{
			{float32(w), float32(-h), float32(gfZ + gfOffset)},
			{float32(-w), float32(-h), float32(gfZ + gfOffset)},
			{float32(-topW), float32(h), float32(gfZ + gfOffset)},
			{float32(topW), float32(h), float32(gfZ + gfOffset)},
			{float32(w), float32(-h), float32(gfZ)},
			{float32(-w), float32(-h), float32(gfZ)},
			{float32(-topW), float32(h), float32(gfZ)},
			{float32(topW), float32(h), float32(gfZ)},
		}

		addGatefoldPanel(gfMesh, gfVerts,
			"gatefold_back_inner", "gatefold_back_back",
			[3]float32{0, 0, 1}, [3]float32{0, 0, -1},
			trapRatio, false, nil, 0)

	case GatefoldDoubleFront:
		// Two half-width flaps on the front face, splitting at x=0
		// Left flap: from -w to 0 (hinges on left edge at x=-w)
		// Right flap: from 0 to w (hinges on right edge at x=w)
		// Node names: "GatefoldFrontLeft", "GatefoldFrontRight"

		gfZ := d
		gfOffset := gfD

		// Determine the back texture key — use gatefold_double_back if available,
		// otherwise fall back to gatefold_double_inner
		backTex := "gatefold_double_back"
		if _, ok := atlas.Positions[backTex]; !ok {
			backTex = "gatefold_double_inner"
		}

		// --- Left flap ---
		gfLeftMesh := &MeshPart{Name: "GatefoldFrontLeft"}
		parts = append(parts, gfLeftMesh)

		gfLeftVerts := [][3]float32{
			{float32(-w), float32(-h), float32(gfZ + gfOffset)},      // 0: bottom-left front
			{float32(0), float32(-h), float32(gfZ + gfOffset)},       // 1: bottom-center front
			{float32(0), float32(h), float32(gfZ + gfOffset)},        // 2: top-center front
			{float32(-topW), float32(h), float32(gfZ + gfOffset)},    // 3: top-left front
			{float32(-w), float32(-h), float32(gfZ)},                 // 4: bottom-left back
			{float32(0), float32(-h), float32(gfZ)},                  // 5: bottom-center back
			{float32(0), float32(h), float32(gfZ)},                   // 6: top-center back
			{float32(-topW), float32(h), float32(gfZ)},               // 7: top-left back
		}

		addGatefoldPanel(gfLeftMesh, gfLeftVerts,
			"gatefold_front_left", backTex,
			[3]float32{0, 0, 1}, [3]float32{0, 0, -1},
			trapRatio, false, nil, 0)

		// --- Right flap ---
		gfRightMesh := &MeshPart{Name: "GatefoldFrontRight"}
		parts = append(parts, gfRightMesh)

		gfRightVerts := [][3]float32{
			{float32(0), float32(-h), float32(gfZ + gfOffset)},       // 0: bottom-center front
			{float32(w), float32(-h), float32(gfZ + gfOffset)},       // 1: bottom-right front
			{float32(topW), float32(h), float32(gfZ + gfOffset)},     // 2: top-right front
			{float32(0), float32(h), float32(gfZ + gfOffset)},        // 3: top-center front
			{float32(0), float32(-h), float32(gfZ)},                  // 4: bottom-center back
			{float32(w), float32(-h), float32(gfZ)},                  // 5: bottom-right back
			{float32(topW), float32(h), float32(gfZ)},                // 6: top-right back
			{float32(0), float32(h), float32(gfZ)},                   // 7: top-center back
		}

		addGatefoldPanel(gfRightMesh, gfRightVerts,
			"gatefold_front_right", backTex,
			[3]float32{0, 0, 1}, [3]float32{0, 0, -1},
			trapRatio, false, nil, 0)

	case GatefoldFrontAndBack:
		// Full-width flap on front — node name: "GatefoldFront"
		gfFrontMesh := &MeshPart{Name: "GatefoldFront"}
		parts = append(parts, gfFrontMesh)

		gfZ := d
		gfOffset := gfD
		gfFrontVerts := [][3]float32{
			{float32(-w), float32(-h), float32(gfZ + gfOffset)},
			{float32(w), float32(-h), float32(gfZ + gfOffset)},
			{float32(topW), float32(h), float32(gfZ + gfOffset)},
			{float32(-topW), float32(h), float32(gfZ + gfOffset)},
			{float32(-w), float32(-h), float32(gfZ)},
			{float32(w), float32(-h), float32(gfZ)},
			{float32(topW), float32(h), float32(gfZ)},
			{float32(-topW), float32(h), float32(gfZ)},
		}

		// Determine front flap back face flip/rotation based on box type
		flipFrontBack := false
		var flipFrontVert *bool
		frontBackRotation := 0

		if gameInfo.BoxType == models.FindBoxTypeIDByName("Eidos Trapezoid") || gameInfo.BoxType == models.FindBoxTypeIDByName("Small Box With Vertical Gatefold") {
			flipFrontBack = true
			f := false
			flipFrontVert = &f
		} else {
			flipFrontBack = isTrapezoid
		}

		addGatefoldPanel(gfFrontMesh, gfFrontVerts,
			"gatefold_front_inner", "gatefold_front_back",
			[3]float32{0, 0, 1}, [3]float32{0, 0, -1},
			trapRatio, flipFrontBack, flipFrontVert, frontBackRotation)

		// Full-width flap on back — node name: "GatefoldBack"
		gfBackMesh := &MeshPart{Name: "GatefoldBack"}
		parts = append(parts, gfBackMesh)

		gfBackZ := -d
		gfBackOffset := -gfD
		gfBackVerts := [][3]float32{
			{float32(w), float32(-h), float32(gfBackZ + gfBackOffset)},
			{float32(-w), float32(-h), float32(gfBackZ + gfBackOffset)},
			{float32(-topW), float32(h), float32(gfBackZ + gfBackOffset)},
			{float32(topW), float32(h), float32(gfBackZ + gfBackOffset)},
			{float32(w), float32(-h), float32(gfBackZ)},
			{float32(-w), float32(-h), float32(gfBackZ)},
			{float32(-topW), float32(h), float32(gfBackZ)},
			{float32(topW), float32(h), float32(gfBackZ)},
		}

		addGatefoldPanel(gfBackMesh, gfBackVerts,
			"gatefold_back_inner", "gatefold_back_back",
			[3]float32{0, 0, 1}, [3]float32{0, 0, -1},
			trapRatio, false, nil, 0)
	}

	return parts
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