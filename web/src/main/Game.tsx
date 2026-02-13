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
                <h1 className='text-[32px] font-bold text-center'>{game.title}</h1>
                {/* <div className='sm:h-200 h-100 w-[95%] ml-auto mr-auto relative max-w-6xl'>
                    <SingleGame ga={game} zd={0} showTitle={false} showFooter={true} />
                </div> */}
                <div id="game">
                    <div id="game-details p-5">
                        <div className="desc bg-black/50 p-5 mt-5">
                            {game.description}
                        </div>
                    </div>
                </div>
                <div id="variants">
                    <h2>Box Variants</h2>
                    {game.variants && game.variants.length > 0 && (
                        game.variants.map((v: any) => (
                            <div key={v.id} className='variant'>
                                <SingleGame ga={v} zd={0} showTitle={false} showFooter={false} />
                            </div>
                        ))
                    )}
                </div>
            </>
            }
        </div>
    )
}