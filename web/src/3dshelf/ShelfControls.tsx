'use client'

import { MapControls } from "@react-three/drei";
import { useEffect, useRef, useState } from "react";
import { useStore } from "@/lib/Store";
import { MOUSE, TOUCH } from 'three';
import { find, forEach, has, isEmpty, isEqual, sample } from "lodash";
// import { useParams } from "next/navigation";
import gsap from 'gsap';
import { useShelvesData } from "./ShelvesProvider";
import { useParams } from "react-router";

export default function ShelfControls()
{
    // @ts-ignore
    const mainControls = useRef<MapControls>(null);
    const {goToSearchedGame,activeShelves,setGoToSearchedGame} = useStore();
    const params = useParams()
    const [controlsEnable, setControlsEnable] = useState(true)
    const startingCameraZ = 30
    const context = useShelvesData()
    const prevShelvesRef = useRef(activeShelves); // Track previous shelves
    
    const { shelfLength } = context;

    useEffect(() => {
        if(isEmpty(params))
        {
            setControlsEnable(true)
        }
        else
        {
            setControlsEnable(false)
        }
    },[params])

    useEffect(() => {
        if(!isEmpty(goToSearchedGame))
        {
            gotoGame(searchShelves({id:goToSearchedGame.variant_id}),false)
            setGoToSearchedGame(null)
        }
    }, [goToSearchedGame]);

    useEffect(() => {
        if(activeShelves && activeShelves.length && !isEqual(prevShelvesRef.current, activeShelves))
        {
            prevShelvesRef.current = activeShelves; // Update ref
            
            let g:any
            if(has(params,'gameSlug') && params['gameSlug'] && params['gameSlug'].length)
            {
                g = searchShelves({'slug':params.gameSlug})
            }
            if(g === undefined)
            {
                g = searchShelves(false)
            }
            gotoGame(g,true)
        }
    },[activeShelves]);

    const searchShelves = function(q:any){
        let shelfCount = 0
        let games:any = []
        forEach(activeShelves, function(s)
        {
            let shelfY = s.shelfY
            forEach(s.games, function(g){
                games.push({
                    id: g.id,
                    slug: g.slug,
                    shelfX: g.shelfX,
                    shelfY: shelfY+(g.h/2),
                    shelfNum: shelfCount
                })
            })
            shelfCount++
        })

        if(q === false)
        {
            return sample(games)
        }
        else
        {
            return find(games,q)
        }
    }

    const gotoGame = function(g:any,firstLoad:boolean){
        let controls = mainControls.current;
        let pos = {
            x: g.shelfX - shelfLength/2,
            y: g.shelfY
        }

        if(!firstLoad)
        {
            // ANIMATE
            gsap.to(controls.target, {
                x: pos.x,
                y: pos.y,
                z: -4,
                duration: 1.2,
                ease: "power2.inOut",
                onUpdate: () => {
                    controls.update(); // Important: update controls during animation
                }
            });
            
            // Animate the camera position
            gsap.to(controls.object.position, {
                x: pos.x,
                y: pos.y,
                z: startingCameraZ,
                duration: 1.2,
                ease: "power2.inOut",
                onUpdate: () => {
                    controls.update(); // Important: update controls during animation
                }
            });
        }
        else
        {
            // NO ANIMATE
            controls.target.set(pos.x,pos.y,-4)
            controls.object.position.set(pos.x,pos.y, startingCameraZ) // starting camera z position
        }    
    }

    return(
        <>
            <MapControls 
                //enableRotate={false}
                enabled={controlsEnable} 
                mouseButtons={{LEFT: MOUSE.PAN, RIGHT:undefined}} 
                touches={{ONE: TOUCH.PAN, TWO: undefined}}
                minDistance={20} 
                maxDistance={50}
                screenSpacePanning={true} 
                ref={mainControls} 
            />
        </>
    )
}