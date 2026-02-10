import { useEffect, useRef, useState } from "react";
import { useStore } from "@/lib/Store";
import Select from 'react-select'
import { useShelvesData } from "./ShelvesProvider";
import { isNumber } from "lodash";
import { useParams } from "react-router";

export default function ShelfOptions() {
    const {setStagedOptions, stagedOptions} = useStore();
    const [showOptions, setShowOptions] = useState(false)
    const [boxTypes, setBoxTypes] = useState([])
    const [devs,setDevs] = useState<Array<any>>([])
    const [pubs,setPubs] = useState<Array<any>>([])
    const [activeDev,setActiveDev] = useState<any>(null)
    const [activePub,setActivePub] = useState<any>(null)
    const context = useShelvesData()
    
    const { publishers, developers, setShelfLength } = context;
    const params = useParams()
    
    const lengthRef = useRef<HTMLInputElement>(null)

    useEffect(() => {
        if(typeof developers !== 'undefined' && devs.length === 0)
        {
            let ds:Array<any> = []
            developers.forEach((d:any) => {
                if(d.count > 0)
                {
                    ds.push({value:d.id,label:d.name+' ('+d.count+')'})
                }
            });
            setDevs(ds)
        }
    },[developers])

    useEffect(() => {
        if(typeof publishers !== 'undefined' && pubs.length === 0)
        {
            let ps:Array<any> = []
            publishers.forEach((d:any) => {
                if(d.count > 0)
                {
                    ps.push({value:d.id,label:d.name+' ('+d.count+')'})
                }
            });
            setPubs(ps)
        }
    },[publishers])

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
            boxTypes:boxTypes,
            dev: activeDev?.value || null, 
            pub: activePub?.value || null
        });
    }

    function clearSettings()
    {
        setBoxTypes([])
        setShelfLength(100)
        setActiveDev(null)
        setActivePub(null)

        setStagedOptions({
            boxTypes:[],
            dev:null,
            pub:null
        });
    }
    
    return (
        !(params && params.gameSlug) && <div id="shelfControls" className="text-sm">
            <a href="#" onClick={() => setShowOptions(!showOptions)} className="text-center text-white my-2 block"> vv Show Options vv</a>
            {showOptions && <div><div id="shelfOptions" className="overflow-y-scroll h-[400px]">
                    
                    {developers && <div className="mt-3">
                        <label htmlFor="" className="text-white">Developer</label>
                        <Select 
                            options={devs} 
                            onChange={(option:any) => {if(option){setActiveDev(option);}}} 
                            isClearable={true}
                            isDisabled={isNumber(stagedOptions.pub)}
                            value={activeDev}
                         />
                    </div>}
                    {publishers && <div className="mt-3">
                        <label htmlFor="" className="text-white">Publisher</label>
                        <Select 
                            options={pubs} 
                            onChange={(option:any) => {if(option){setActivePub(option);}}} 
                            isClearable={true} 
                            isDisabled={isNumber(stagedOptions.dev)}
                            value={activePub}
                        />
                    </div>}
                    <div className="mt-3">
                        <label htmlFor="" className="text-white">Shelf Length</label>
                        <input type="number" ref={lengthRef} defaultValue={100} className='w-full border-white border-1 text-black block p-1 mt-2 mb-2 drop-shadow-xl text-white' />
                    </div>
                </div>
                <a href="#" onClick={() => applySettings()} className="text-black text-center bg-white inline-block border-white-1 border-1 p-2 my-5">Apply Settings</a>
                <a href={"/shelves"} onClick={() => clearSettings()} className="text-white text-center inline-block border-white-1 border-1 p-2 my-5 ml-5">Clear Settings</a>
            </div>}
        </div>
    )
}