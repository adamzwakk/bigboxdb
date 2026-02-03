"use client"

import React, {useEffect, useRef, useState} from "react";
import {useStore} from "@/lib/Store";
import {forEach} from "lodash";
import type { SearchIndex } from "@/lib/types";
import { useParams } from "react-router";

export default function Search({onShelf}:{onShelf:boolean})
{
    const { setShouldHover, setGoToSearchedGame, goToSearchedGame } = useStore();
    const [searchResults, setSearchResults] = useState<Array<SearchIndex>>()
    const [showSearch, setShowSearch] = useState(true)
    const searchRef = useRef<HTMLInputElement>(null)
    const [searchIndex, setSearchIndex] = useState<Array<any>>()
    const params = useParams()

    const search = function(e:React.KeyboardEvent<HTMLElement>)
    {
        let et = e.target as any
        let v = et.value
        if(v.length <= 1)
        {
            setSearchResults([])
            return
        }
        if(!searchIndex)
        {
            fetch('/api/search')
                .then((res) => res.json())
                .then((d) => setSearchIndex(d))
        }
        else
        {
            let found:Array<any> = []
            forEach(searchIndex, function (bb){
                if(bb.title.toLowerCase().includes(v.toLowerCase())){
                    found.push({id:bb.id, slug:bb.slug, name:bb.title, img:'/scans/'+bb.slug+'/front.webp'})
                }
            });
            if(found.length)
            {
                setSearchResults(found);
                if(onShelf)
                {
                    setShouldHover(found)
                }
            }
        }
    }

    useEffect(() => {
        if(!onShelf){return}
        if(!(params && params.gameSlug))
        {
            setShowSearch(true)
        }
        else
        {
            setShowSearch(false);
            setSearchResults([])
        }
    },[params])

    useEffect(() => {
        if(onShelf && goToSearchedGame)
        {
            setSearchResults([])
            if(searchRef.current !== null && typeof searchRef.current !== 'undefined')
            {
                searchRef.current.value = ''
            }
        } else if(!onShelf && goToSearchedGame) {
            location.href = '/game/'+goToSearchedGame.slug
        }
    },[goToSearchedGame])

    return(
        showSearch && <div id="searchBox" className="relative">
            <input onKeyUp={search} type="text" placeholder='Find a box...' className='w-full searchBox text-black block p-1 mt-2 mb-2 drop-shadow-xl bg-white' />
            {searchResults && searchResults.length > 0 && <ul className='searchResults bg-black/80 top-[36px] absolute w-full z-10'>
                {searchResults.map((g:any) => <li className="mb-2 p-2 hover:bg-[#4a4a4a] last:mb-0 text-white" key={g.slug} onClick={() => {setGoToSearchedGame(g)}}>
                    <img src={g.img} className="inline w-5 mr-2" />{g.name}</li>
                )}
            </ul>}
        </div>
    )
}