import {useEffect, useState} from "react";
import {isEmpty, find, merge} from "lodash";
import { BigBoxTypes, NewBigBoxTypes } from "@/lib/enums";
import type { GameInfoProps } from "@/lib/types";
import { useShelvesData } from "./ShelvesProvider";
import { useParams } from "react-router";

function GameInfo(){

    const [showingInfo, setShowingInfo] = useState(false)
    const [showingNotes, setShowingNotes] = useState(false)
    const [cachedInfo,setCachedInfo] = useState<any>([])
    const [activeGame,setActiveGame]= useState<any>({})
    const params = useParams()
    const context = useShelvesData()
    const { allGames } = context;

    useEffect(() => {
        if(allGames && params && params.variantId)
        {
            let vid = parseInt(params.variantId)
            if(!find(allGames,{id:vid}))
            {
                return // dont load anything if you arent a known slug
            }
            let info:GameInfoProps
            info = find(cachedInfo,{id:vid})
            if(info)
            {
                setActiveGame(merge(info,activeGame))
            }
            else
            {
                fetch('/api/variants/'+vid)
                    .then((res) => res.json())
                    .then((d:any) => {
                        let info:GameInfoProps = d
                        info.title = d.title
                        info.width = d.w
                        info.height = d.h
                        info.depth = d.d
                        info.box_type_name = d.box_type_name
                        info.scan_notes = d.scan_notes
                        setActiveGame(info)
                        let a = [...cachedInfo,info]
                        setCachedInfo(a) // save for next time they ask
                    });
            }
        }
        else
        {
            setShowingInfo(false)
            setShowingNotes(false)
            setActiveGame(null)
        }
    },[allGames,params])

    const showInfo = function()
    {
        setShowingInfo(!showingInfo)
    }

    const showNotes = function()
    {
        setShowingNotes(!showingNotes)
    }

    return (
        <>
            <div id="topHud">
                {isEmpty(activeGame) && (
                    <div className="logoContain">
                        <div>
                            <div className="mainLogo">
                                <img src="/img/logo_filled.png" alt="Logo"/>
                            </div>
                            <div className="mainLogoText">
                                <h1 className="font-bold">BigBoxDB</h1>
                                <h2 className="font-bold">an elegant wrapping<br/>from a more civilized age</h2>
                            </div>
                        </div>
                    </div>
                ) || (
                    <div className="logoContainActive">
                        <div>
                            <div className="mainLogo">
                                <img src="/img/logo_filled.png" alt="Logo"/>
                            </div>
                        </div>
                    </div>
                )}
                {activeGame && <div id="gameInfo">
                    <h3 className="font-bold">{activeGame.title} ({activeGame.year})</h3>
                    <div className="showInfoButton" onClick={showInfo}>Toggle More Info</div>
                    {showingInfo && 
                        <div className="innerGameInfo">
                            <table className="info">
                                <tbody>
                                    <tr><td><strong>Platform:</strong></td><td>{activeGame.platform}</td></tr>
                                    <tr><td><strong>Variant:</strong></td><td>{activeGame.variant}</td></tr>
                                    <tr><td><strong>Box Type:</strong></td><td>{BigBoxTypes[activeGame.box_type].replaceAll('_', ' ')}</td></tr>
                                    <tr><td><strong>WxHxD:</strong></td><td>{activeGame.w} x {activeGame.h} x {activeGame.d}</td></tr>
                                    <tr><td><strong>Contributed By:</strong></td><td>{activeGame.contributed_by}</td></tr>
                                    {activeGame.scan_notes && <tr><td><strong>Scan Notes:</strong></td><td><a href="#" onClick={showNotes}>Toggle Notes</a></td></tr>}
                                </tbody>
                            </table>
                            <div className="clear"></div>
                            {(NewBigBoxTypes.has(activeGame.box_type)) && (
                                <strong style={{display:'block'}}>
                                    Newer retro game!
                                </strong>
                            )}
                            {(activeGame.steam_link || activeGame.gog_link || activeGame.other_link) && (
                                <div className="links">
                                    <div className="stillFound">Still found on:</div>
                                    <ul className="actualLinks">
                                        {activeGame.steam_link && <li key='steam'><a className={'steam'} href={activeGame.steam_link} target='_blank'></a></li>}
                                        {activeGame.gog_link && <li key='gog'><a className={'gog'} href={activeGame.gog_link} target='_blank'></a></li>}
                                        {activeGame.other_link && <li key='other'><a className={'other'} href={activeGame.ohter_link} target='_blank'></a></li>}
                                    </ul>
                                </div>
                            )}  
                            {((!activeGame.steam_link && !activeGame.gog_link && !activeGame.other_link) && !NewBigBoxTypes.has(activeGame.box_type)) && (
                                <strong>Probably Abandonware!</strong>
                            )}
                        </div>
                    }
                </div>}
            </div>
            {(activeGame && activeGame.scan_notes && showingNotes) && <div id="scanNotes" className="center-ui-element z-99">
                <div className="innerNotes">
                    {activeGame.scan_notes.split('\n').map(function(item:any, key:any) {
                        return (
                            <span key={key}>
                            {item}
                            <br/>
                            </span>
                        )
                    })}
                </div>
                {activeGame.scan_notes.split('\n').length > 3 && <div className="buttons">
                    <a href="#" onClick={() => (setShowingNotes(false))}>Close Notes</a>
                </div>}
            </div>}
        </>
    )
}

export default GameInfo