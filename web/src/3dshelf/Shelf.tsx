import BigBox from "./BigBox";
import { SRGBColorSpace } from "three";
import { useTexture } from "@react-three/drei";
import type { Game3D } from "@/lib/types";

interface ShelfProps{
    games: Game3D[],
    shelfY: number
    width: number
}

function Shelf({games, shelfY, width}: ShelfProps){
    const height = 1, depth = 12
    
    const texture = useTexture('/img/3dshelf/wood.jpg', function(tex){
        tex.colorSpace = SRGBColorSpace
    })
    
    return (
        <group>
            <mesh position={[0, shelfY, 0]} onPointerEnter={(e) => e.stopPropagation()}>
                <boxGeometry args={[width, height, depth]}/>
                <meshStandardMaterial map={texture} />
            </mesh>
            {games.map((g, index) => {            
                return (    
                    <BigBox 
                        key={index} 
                        position={{
                            x: g.shelfX! - width/2,
                            y: shelfY + height/2 + g.h!/2,
                            z: g.shelfZ! + depth/2
                        }} 
                        g={g}
                        onShelf={true}
                    />
                )
            })}
        </group>
    )
}

export default Shelf;