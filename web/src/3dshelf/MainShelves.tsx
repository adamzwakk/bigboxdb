import '../globals.css'
import './3dshelf.scss'

import type { Game3D, ShelfProps } from "@/lib/types";
import { filter, forEach, map } from "lodash";
import { BoxShelfDirection } from "@/lib/enums";
import { useStore } from '@/lib/Store';

import { useCallback, useEffect, useRef, useState } from 'react';
import BackBlur from './BackBlur';
import { Canvas } from '@react-three/fiber';
import { AdaptiveDpr, AdaptiveEvents, Bvh } from '@react-three/drei';
import ShelfControls from './ShelfControls';
import Shelf from './Shelf';
import { Group } from 'three';
import { useShelvesData } from './ShelvesProvider';
import { useLocation, useParams } from 'react-router';

export default function MainShelves() {
    const {setShelfNum, activeShelves, setActiveShelves, stagedOptions} = useStore();
    const padding: {x:number, z:number} = {x:2, z:0.5}
    const allShelves = useRef<Group>(null);
    const pathname = useLocation().pathname;
    const params = useParams()
    const context = useShelvesData()

    const { allGames, shelfLength, setShelfLength } = context;
    
    // Add local state to track the current shelf length for rendering
    const [currentShelfLength, setCurrentShelfLength] = useState(shelfLength);

    useEffect(() => {
        if(allGames && allGames.length) {
            rerenderShelves()
        }
    },[allGames,stagedOptions])

    const filterGames = useCallback(function() {
        let g:any
        if(allGames.length)
        {
            let games = structuredClone(allGames)

            if(pathname.includes('/developer/') && params.devSlug)
            {
                let dslug = params.devSlug
                games = filter(games, (e) => {
                    return e.developers?.some((d: {name: string, slug: string}) => {
                        return d.slug == dslug;
                    });
                });
            }
            else if(pathname.includes('/publisher/') && params.pubSlug)
            {
                let pslug = params.pubSlug
                games = filter(games, (e) => {
                    return e.publishers?.some((d: {name: string, slug: string}) => {
                        return d.slug == pslug;
                    });
                });
            }

            // if(stagedOptions.boxTypes.length)
            // {
            //     games = filter(games, function(o) {  return stagedOptions.boxTypes.includes(o.box_type); });
            // }


            if(games)
            {
                return games
            }
        }

        return allGames
    }, [allGames, pathname, params, stagedOptions]);

    const rerenderShelves = useCallback(function () {
        let filteredGames = filterGames()
        
        const result = getShelves(filteredGames, shelfLength)  // Pass games and shelfLength separately
        if(result.shelves[0]?.games.length)
        {
            setActiveShelves(result.shelves)
            setShelfNum(result.shelves.length)
            setShelfLength(result.finalShelfLength)
            setCurrentShelfLength(result.finalShelfLength)
        }
    }, [shelfLength, allGames, stagedOptions, filterGames]);  // Add filterGames to dependencies

    function getShelves(games:Game3D[], initialShelfLength:number): { shelves: ShelfProps[], finalShelfLength: number } {
        const cleaned = structuredClone(games);
        let activeWidth:number = 0
        let shelfGames:Array<Game3D> = [];
        let currentShelfMaxHeight:number = 0
        let activeShelf:Array<{games: Game3D[], maxHeight: number, width: number}> = []  // Added width here
        let shelfX:number = padding.x;
        let lastBox:Game3D;
        
        // Calculate shelf length first
        let shelfLength = initialShelfLength;
        let totalWidth = 0;
        forEach(cleaned, function(g:Game3D) {
            g.sd = g.dir === BoxShelfDirection.front ? g.w : g.d;
            totalWidth += g.sd;
        });
        
        if(totalWidth < initialShelfLength - padding.x) {
            shelfLength = totalWidth + cleaned[cleaned.length-1].sd + (padding.x/2);
        }

        forEach(cleaned, function(g:Game3D)
        {
            const gameSeed = g.slug;            
            let chance:boolean = lastBox && lastBox.dir !== BoxShelfDirection.front && Math.random() < 0.10;
            if(chance && g.worth_front_view === true && (activeWidth+g.w) < shelfLength-2)
            {
                g.dir = BoxShelfDirection.front
            }

            const zHash = hashString(gameSeed + '_z');
            let zVariance:number = parseFloat(((zHash % 1000 / 1000) * (0.320 - 0.100) + 0.100).toFixed(4));

            if(g.dir == BoxShelfDirection.left)
            {
                g.shelfZ = (padding.z - zVariance) - g.w/2 - 0.8;
                g.sd = g.d

                if(lastBox){
                    shelfX = lastBox.shelfX! + (lastBox.sd/2) + (g.d/2);
                }
            }
            else
            {
                g.shelfZ = (padding.z - zVariance) - g.d/2 - 0.8;
                g.sd = g.w

                if(lastBox){
                    shelfX = lastBox.shelfX! + (lastBox.sd/2) + (g.w/2);
                }
                else
                {
                    shelfX = shelfX + (g.w/2);
                }
            }

            activeWidth += g.sd;
            g.shelfX = shelfX;
            lastBox = g;
            if(activeWidth < shelfLength-padding.x || cleaned.length === 1)
            { 
                if(cleaned.length === 1)
                {
                    g.dir = BoxShelfDirection.front
                    shelfLength = padding.x + g.w + padding.x
                    g.shelfX = g.w/2 + g.w/4
                }
                shelfGames.push(g);
                // Update max height ONLY when game is added to shelf
                if(g.h! > currentShelfMaxHeight)
                {
                    currentShelfMaxHeight = g.h
                }
            }
            else
            {
                // Store the CURRENT shelf before moving to next
                activeShelf.push({
                    games: shelfGames, 
                    maxHeight: currentShelfMaxHeight,
                    width: shelfLength
                })
                
                g.shelfX = padding.x;
                if(g.dir === BoxShelfDirection.front)
                {
                    g.shelfX += g.sd/2 - padding.x/2
                }
                shelfGames = [g]
                activeWidth = g.sd
                currentShelfMaxHeight = g.h  // NEW shelf starts with this game's height
                shelfX = g.shelfX;
            }
        });
        
        // Add final shelf with its width
        activeShelf.push({
            games: shelfGames, 
            maxHeight: currentShelfMaxHeight,
            width: shelfLength
        })

        let cumulativeY = 0;
        const finalShelves: ShelfProps[] = activeShelf.map((shelf) => {
            const shelfYPosition = cumulativeY;
            cumulativeY -= (shelf.maxHeight + 1.5);
            
            return {
                games: shelf.games,
                shelfY: shelfYPosition,
                width: shelf.width  // Use the width stored with this shelf
            };
        });
        
        if(isNaN(shelfLength))
        {
            // just so its SOMETHING
            shelfLength = 10
        }

        return { shelves: finalShelves, finalShelfLength: shelfLength }
    }

    // Add this helper function at the top of your component
    function hashString(str: string): number {
        let hash = 0;
        for (let i = 0; i < str.length; i++) {
            hash = ((hash << 5) - hash) + str.charCodeAt(i);
            hash = hash & hash;
        }
        return Math.abs(hash);
    }

    return (
        <div id="roomScene">
            <Canvas performance={{ min: 0.2, max:0.5 }} gl={{ 
                    powerPreference: "high-performance",
                }}
                onCreated={({ gl }) => {
                    gl.setClearColor(0x000000, 0);
                }} shadows={false}>
                <AdaptiveDpr pixelated />
                <AdaptiveEvents />
                <Bvh>
                    <ShelfControls />
                    <ambientLight intensity={1.6} />
                    <group ref={allShelves}>
                        {activeShelves.map((s:any, index:number) => <Shelf shelfY={s.shelfY} key={index} games={s.games} width={currentShelfLength} /> )}
                    </group>
                    <BackBlur />
                </Bvh>
            </Canvas>
        </div>
    );
}