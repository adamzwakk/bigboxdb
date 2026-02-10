import {useEffect, useRef, useState} from "react";
import {useStore} from "@/lib/Store";
import { useNavigate, useParams } from "react-router";
import { InstantSearch, SearchBox, Hits, useInstantSearch, useSearchBox, Configure } from 'react-instantsearch';
import { instantMeiliSearch } from '@meilisearch/instant-meilisearch';

export default function Search({onShelf}:{onShelf:boolean})
{
    const navigate = useNavigate();

    const { searchClient } = instantMeiliSearch(
        import.meta.env.VITE_MEILI_URL, 
        import.meta.env.VITE_MEILI_KEY
    );
    
    const { setGoToSearchedGame, goToSearchedGame } = useStore();
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
        }
    },[params, onShelf])
    
    useEffect(() => {
        if(onShelf && goToSearchedGame)
        {
            if(searchRef.current !== null && typeof searchRef.current !== 'undefined')
            {
                searchRef.current.value = ''
            }
        } else if(!onShelf && goToSearchedGame) {
            navigate('/game/'+goToSearchedGame.slug)
        }
    },[goToSearchedGame, onShelf])
    
    function ConditionalHits({ hitComponent }: { hitComponent: any }) {
        const { results, indexUiState } = useInstantSearch();
        const query = indexUiState.query;
        
        if (!query || !results || results.nbHits === 0) return null;
        
        return <Hits hitComponent={hitComponent} classNames={{list: 'searchResults bg-black/80 top-[36px] absolute w-full z-10'}} />;
    }
    
    function Hit({ hit }:{hit:any}) {
        const { refine } = useSearchBox(); // Add this hook
        
        const handleClick = () => {
            setGoToSearchedGame(hit);
            refine('')
        }
        
        return (
            <div className="mb-2 w-full block p-2 hover:bg-[#4a4a4a] last:mb-0 text-white cursor-pointer" onClick={handleClick}>
                <img src={'/scans/'+hit.slug+'/'+hit.variant_id+'/front.webp'} className="inline w-5 mr-2" alt={hit.title} />
                <span className="title">{hit.title}</span>
                <span className="year inline-block ml-2 text-xs text-[#7b7b7b]">({hit.year})</span>
            </div>
        );
    }
    
    return(
        showSearch && <div id="searchBox" className="relative">
            <InstantSearch future={{preserveSharedStateOnUnmount:true}} indexName="items" searchClient={searchClient}>
                <Configure hitsPerPage={5} />
                <SearchBox 
                    placeholder="Search boxes..." 
                    classNames={{
                        submitIcon:'hidden', 
                        resetIcon:'hidden', 
                        input:'w-full searchBox text-black block p-1 drop-shadow-xl bg-white', 
                        submit:'hidden'
                    }} 
                />
                <ConditionalHits hitComponent={Hit} />
            </InstantSearch>
        </div>
    )
}