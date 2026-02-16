import Header from '@/partials/Header'
import '@/globals.css'
import '@/main/main.scss'
import SingleGame from '@/partials/SingleGame';
import { useEffect, useState } from 'react';
import { useParams } from 'react-router';
import type { Game3D } from '@/lib/types';

export default function Variant() {
    const params = useParams()
    const id = parseInt(params.variantId!)
    const [game,setGame] = useState<Game3D>()

    useEffect(() => {
        fetch('/api/variants/'+id)
            .then(res => res.json())
            .then((data) => {
                setGame(data)
                document.title = `${data.title} (${data.variant}) | BigBoxDB`;
            })
            .catch(console.error)
    },[])

    return(
        <div className='relative z-4 text-white w-full ml-auto mr-auto max-w-4xl'>
            <Header />
            {game && <>
                <h1 className='text-[32px] font-bold text-center'>
                    <a className='underline' href={"/game/"+game.game_slug}>{game.title}</a> ({game.variant} {game.box_type_name})
                </h1>
                <div className='sm:h-200 h-100 w-[95%] ml-auto mr-auto relative max-w-6xl'>
                    <SingleGame ga={game} zd={0} showTitle={false} showFooter={true} />
                </div>
                <div id="game">
                    <div id="game-details p-5">
                        <div className="desc bg-black/50 p-5 mt-5">
                            <ul>
                                <li><h3 className='font-bold text-[18px] mb-2 inline'>Release Year:</h3> {game.year}</li>
                                <li><h3 className='font-bold text-[18px] mb-2 inline'>Platform:</h3> {game.platform}</li>
                                <li><h3 className='font-bold text-[18px] mb-2 inline'>Variant:</h3> {game.variant} {game.box_type_name}</li>
                                <li><h3 className='font-bold text-[18px] mb-2 inline'>Dimensions (inches):</h3> {game.w} x {game.h} x {game.d}</li>
                                <li><h3 className='font-bold text-[18px] mb-2 inline'>Contributed By:</h3> {game.contributed_by}</li>
                            </ul>
                        </div>
                        {game.scan_notes && <div className="desc bg-black/50 p-5 mt-5">
                             <div className='mt-2'>
                                <h3 className='font-bold text-[18px] mb-2'>Scan Notes:</h3><p className='ml-0 md:ml-7'>{game.scan_notes}</p>
                            </div>
                        </div>}
                    </div>
                </div>
                </>
            }
        </div>
    )
}