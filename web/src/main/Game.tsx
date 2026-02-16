import Header from '@/partials/Header'
import '@/globals.css'
import '@/main/main.scss'
import { useEffect, useState } from 'react';
import { useParams } from 'react-router';

export default function Game() {
    const params = useParams()
    const slug = params.gameSlug!
    const [game,setGame] = useState<any>()

    useEffect(() => {
        fetch('/api/games/'+slug)
            .then(res => res.json())
            .then((data) => {
                setGame(data)
                document.title = `${data.title} | BigBoxDB`;
            })
            .catch(console.error)
    },[])

    return(
        <div className='relative z-4 text-white w-full ml-auto mr-auto max-w-4xl'>
            <Header />
            {game && <>
                <h1 className='text-[32px] font-bold'>{game.title} ({game.platform})</h1>
                <div id="game" className='flex mt-5 gap-5'>
                    <div className="w-md flex-1">
                        <img src={"/scans/"+game.slug+'/'+game.variants[0].id+'/front.webp'} alt="" />
                    </div>
                    <div className="flex-2">
                        <div className="bg-black/50 p-5">
                            {game.description}
                        </div>
                        <div className="bg-black/50 p-5 mt-5">
                            <h4 className='font-bold'>Links:</h4>
                            <ul className='gameLinks mt-2'>
                                {game.mobygames_id && game.mobygames_id > 0 && <li className='inline'><a href={"https://www.mobygames.com/game/"+game.mobygames_id} target='_blank' className={"mobygames w-[32px] h-[32px] bg-no-repeat bg-size-[100%] inline-block mr-1"}></a></li>}
                                {game.igdb_slug && game.igdb_slug > 0 && <li className='inline'><a href={"https://www.igdb.com/game/"+game.igdb_slug} target='_blank' className={"igdb w-[32px] h-[32px] bg-no-repeat bg-size-[100%] inline-block mr-1"}></a></li>}
                                {game.links && game.links.length > 0 && 
                                    game.links.map((l: any) => (
                                        <li key={l.id} className='inline'>
                                            <a href={l.link} className={l.name+" w-[32px] h-[32px] bg-no-repeat bg-size-[100%] inline-block mr-2"} target='_blank'></a>
                                        </li>
                                    ))
                                }
                            </ul> 
                        </div>
                    </div>
                </div>
                {game.variants && game.variants.length > 0 && <div className='mt-5'>
                    <h2 className='text-[24px] font-bold my-5'>Editions ({game.variants.length})</h2>
                    <div className="flex justify-start gap-10">
                        {(game.variants.map((v: any) => (
                            <a href={"/game/"+game.slug+"/"+v.id} key={v.id} className='variant w-[25%] bg-black/50 p-5'>
                                <div>{v.name}</div>
                                <img src={"/scans/"+game.slug+"/"+v.id+"/front.webp"} alt="" className='w-[100%]' />
                                <ul>
                                    <li>{v.box_type_name}</li>
                                </ul>
                                {/* <SingleGame ga={v} zd={0} showTitle={false} showFooter={false} /> */}
                            </a>
                        )))}
                    </div>
                </div>
                }
            </>
            }
        </div>
    )
}