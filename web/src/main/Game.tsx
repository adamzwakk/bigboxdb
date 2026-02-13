import Header from '@/partials/Header'
import '@/globals.css'
import '@/main/main.scss'
import SingleGame from '@/partials/SingleGame';
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
                <h1 className='text-[32px] font-bold'>{game.title}</h1>
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
                            <ul>
                                {game.links && game.links.length > 0 && 
                                    game.links.map((l: any) => (
                                        <li key={l.id} className='inline'>
                                            <a href={l.link} className={l.name}></a>
                                        </li>
                                    ))
                                }
                            </ul> 
                        </div>
                    </div>
                </div>
                <div className='mt-5'>
                    <h2 className='text-[24px] font-bold'>Box Variants</h2>
                    <div className="flex justify-start">
                        {game.variants && game.variants.length > 0 && (
                            game.variants.map((v: any) => (
                                <div key={v.id} className='variant w-[25%] h-50'>
                                    <div>{v.name}</div>
                                    <SingleGame ga={v} zd={0} showTitle={false} showFooter={false} />
                                </div>
                            ))
                        )}
                    </div>
                </div>
            </>
            }
        </div>
    )
}