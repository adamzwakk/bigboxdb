"use client"

import React, {useEffect, useRef, useState} from "react";
import {useStore} from "@/lib/Store";
import {forEach} from "lodash";
import type { SearchIndex } from "@/lib/types";
import { useParams } from "react-router";

import { InstantSearch, SearchBox, Hits } from 'react-instantsearch';
import { instantMeiliSearch } from '@meilisearch/instant-meilisearch';

export default function Search({onShelf}:{onShelf:boolean})
{
    const { searchClient } = instantMeiliSearch(
        import.meta.env.VITE_MEILI_URL, 
        import.meta.env.VITE_MEILI_KEY
    );
    const { setShouldHover, setGoToSearchedGame, goToSearchedGame } = useStore();
    const [showSearch, setShowSearch] = useState(true)
    const searchRef = useRef<HTMLInputElement>(null)
    const params = useParams()

    useEffect(() => {
        if(!onShelf){return}
        if(!(params && params.gameSlug))
        {
            setShowSearch(true)
        }
        else
        {
            setShowSearch(false);
            // setSearchResults([])
        }
    },[params])

    useEffect(() => {
        if(onShelf && goToSearchedGame)
        {
            // setSearchResults([])
            if(searchRef.current !== null && typeof searchRef.current !== 'undefined')
            {
                searchRef.current.value = ''
            }
        } else if(!onShelf && goToSearchedGame) {
            location.href = '/game/'+goToSearchedGame.slug
        }
    },[goToSearchedGame])

    function Hit({ hit }:{hit:any}) {
        console.log(hit.variant_id);

        return (
            
            <div className="search-result" key={hit.variant_id}>
                <h3>{hit.title}</h3>
            </div>
        );
    }

    return(
        showSearch && <div id="searchBox" className="relative">
            <InstantSearch indexName="items" searchClient={searchClient}>
                <SearchBox placeholder="Search products..." />
                <Hits hitComponent={Hit} />
            </InstantSearch>
        </div>
    )
}