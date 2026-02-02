"use client"

import { useEffect, useRef, useState } from "react";
import { useStore } from "@/lib/Store";
import type { StatsProps } from "@/lib/types";
import Select from 'react-select'
import { useShelvesData } from "./ShelvesProvider";
import { useParams } from "react-router";

export default function ShelfOptions() {
    const {stagedOptions, setStagedOptions} = useStore();
    const [showOptions, setShowOptions] = useState(false)
    const [boxTypes, setBoxTypes] = useState([])
    const [stats,setStats] = useState<Array<StatsProps>>()

    const [developer, setDeveloper] = useState<string>()
    const [publisher, setPublisher] = useState<string>()
    const context = useShelvesData()
    
    const { publishers, developers, setShelfLength } = context;
    const params = useParams()
    
    const lengthRef = useRef<HTMLInputElement>(null)

    useEffect(() => {
        let tempStats:any = [];
        fetch('/api/stats')
            .then((res) => res.json())
            .then((s:Array<StatsProps>) => {
                s.forEach(st => {
                    if(typeof st.id !== 'undefined' && st.count > 0 && st.id !== -1)
                    {
                        tempStats.push({value:st.id,label:st.name})
                    }
                });
                setStats(tempStats)
            })

        if(params && params.pubSlug)
        {
            setPublisher(params.pubSlug.toString())
        }

        if(params && params.devSlug)
        {
            setDeveloper(params.devSlug.toString())
        }
    },[])

    function toggleTypes(v:any,a:any)
    {
        if((a.action == 'remove-value' || a.action == 'clear') && v.length === 0){
            setBoxTypes([])
        }
        let newTypes:any = []
        v.forEach((element:any) => {
            let id = element.value
            if(stagedOptions.boxTypes.includes(id))
            {
                newTypes = newTypes.filter((e:any) => e !== id)
            }
            else
            {
                newTypes.push(id)
            }
        });
        setBoxTypes(newTypes)
    }

    function applySettings()
    {
        if(lengthRef.current)
        {
            let num = parseInt(lengthRef.current.value)
            if(num > 15 && num < 500){
                // l = num
                setShelfLength(num)
            }
            if(Number.isNaN(num))
            {
                setShelfLength(100)
            }
        }
        setStagedOptions({
            boxTypes:boxTypes
        });
    }

    useEffect(() => {
        applySettings()
    },[publisher,developer]);
    
    return (
        <div id="shelfControls" className="text-sm">
            <a href="#" onClick={() => setShowOptions(!showOptions)} className="text-center text-white my-2 block"> vv Show Options vv</a>
            {showOptions && <div id="shelfOptions" className="overflow-y-scroll h-[400px]">
                <div className="mt-3">
                    <label htmlFor="" className="text-white">Shelf Length (default 100)</label>
                    <input type="number" ref={lengthRef} className='w-full border-white border-1 text-black block p-1 mt-2 mb-2 drop-shadow-xl text-white' />
                </div>
                {stats && <div className="mt-3">
                    <label htmlFor="" className="text-white">Box Type</label>
                    <Select
                        isMulti 
                        options={stats} 
                        onChange={toggleTypes}
                        className="basic-multi-select"
                    />
                </div>}
                {developers && <div className="mt-3">
                    <label htmlFor="" className="text-white">Developer <a onClick={() => setDeveloper('')} className="text-white underline" href="/shelves">(Clear)</a></label>
                    <ul className="h-[200px] overflow-y-scroll border-1 border-white">
                        {developers.map((s:any, index:number) => <li key={index}><a onClick={() => {setDeveloper(s.slug); setPublisher(''); }} className={(developer == s.slug ? "bg-white text-black" : "text-white" )+" py-1 px-2 block border-b-1 w-full border-white"} href={"/shelves/developer/"+s.slug} dangerouslySetInnerHTML={{__html:s.name+' ('+s.count+')'}}></a></li> )}
                    </ul>
                </div>}
                {publishers && <div className="mt-3">
                    <label htmlFor="" className="text-white">Publisher <a onClick={() => setPublisher('')} className="text-white underline" href="/shelves">(Clear)</a></label>
                    <ul className="h-[200px] overflow-y-scroll border-1 border-white">
                        {publishers.map((s:any, index:number) => <li key={index}><a onClick={() => {setPublisher(s.slug); setDeveloper(''); }} className={(publisher == s.slug ? "bg-white text-black" : "text-white" )+" py-1 px-2 block border-b-1 w-full border-white"} href={"/shelves/publisher/"+s.slug} dangerouslySetInnerHTML={{__html:s.name+' ('+s.count+')'}}></a></li> )}
                    </ul>
                </div>}
                <a href="#" onClick={() => applySettings()} className="text-black text-center bg-white inline-block border-white-1 border-1 p-2 my-5">Apply Settings</a>
            </div>}
        </div>
    )
}