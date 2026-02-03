'use client'

import type { Game3D } from "@/lib/types"
import { Canvas } from "@react-three/fiber"
import { IsTouchDevice } from "@/lib/Utils"
import BigBox from "@/3dshelf/BigBox"
import { useEffect, useState } from "react"

type RandomGameType = {
    slug:boolean|string,
    zd:number,
    showFooter:boolean
}

export default function SingleGame({slug,zd,showFooter}:RandomGameType)
{
    const [g,setG] = useState<Game3D>()
    if(zd == 0)
    {
        zd = IsTouchDevice() ? -6 : -4
    }

    useEffect(() => {
        if(!slug){
            fetch('/api/variants/botd')
                .then((res) => res.json())
                .then((ga:Game3D) => {
                    setG(ga)
                })
        }
        else
        {
            fetch('/api/variants/'+slug)
                .then((res) => res.json())
                .then((ga:Game3D) => {
                    setG(ga)
                })
        }
    },[])

    return(
        <>
            {g && !slug && <a href={'/game/'+g.slug} className='text-[18px] block underline text-center font-bold'>{g.title} ({g.year})</a>}
            <div className='relative h-[95%] overflow-hidden'>
                <Canvas performance={{ min: 0.2, max:0.5 }}>
                    <ambientLight intensity={2} />
                    <directionalLight color={'#ffffff'} intensity={1.2} position={[0,0,10]} target-position={[0,0,0]} />
                    {g && <BigBox g={g} position={{x:0,y:0,z:zd}} onShelf={false} />}
                </Canvas>
            </div>
            {showFooter && g && <div id="threedee-controls" className="hidden sm:block text-center bg-black/50 absolute bottom-[5%] left-[50%] -translate-x-1/2 p-2 rounded-lg">
                {/* <div className="inline"><span className="font-bold">Drag:</span> Rotate</div>&nbsp;
                <div className="inline"><span className="font-bold">Ctrl-Drag:</span> Move</div>
                {AllGatefoldTypes.has(g.box_type) && <div className="inline">&nbsp;<span className="font-bold">D-Click:</span> Open</div>} */}
                <div className="text-center">
                    <img src="/img/icons/shelves.png" className="inline w-6 mr-1" alt="" />
                    <a className="underline" href={"/shelves/game/"+g.slug}>See it on the 3D shelves!</a>
                    <img src="/img/icons/shelves.png" className="inline w-6 ml-1" alt="" />
                </div>
            </div>}
        </>
    )
}