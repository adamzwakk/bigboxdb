import { useEffect, useState } from 'react'
import '../globals.css'
import './main.scss'
import HomeBlurb from '@/partials/HomeBlurb'
import Search from '@/partials/Search'
import SingleGame from '@/partials/SingleGame'
import type { Game3D } from '@/lib/types'

export default function Home() {

    const [latestGames,setLatestGames] = useState([])
    const [botd,setBOTD] = useState<Game3D>()

    useEffect(() => {
        fetch('/api/variants/latest')
            .then(res => res.json())
            .then((data) => {
                setLatestGames(data)
            })
            .catch(console.error)
        
        fetch('/api/variants/botd')
            .then((res) => res.json())
            .then((ga:Game3D) => {
                setBOTD(ga)
            })
    },[])

    return(
      <div id="home" className='relative'>
        <div id="main-header" className="ml-auto mr-auto w-full sm:w-xl p-5 z-4 text-white pt-[3vh]">
            <a href="/" className='mb-5 block'>
                <div className='flex mr-auto ml-auto text-center content-center justify-center'>
                  <img className='mainLogo inline w-[50px] mr-3 mb-3' src="/img/logo_filled.png" alt="Logo"/>
                  <h1 className='text-[50px] text-center block font-bold'>BigBoxDB</h1>
                </div>
                <h2 className='text-[18px] -mt-2 text-center block font-bold'>an elegant wrapping from a more civilized age</h2>
            </a>
            <Search onShelf={false} />
            <ul className='text-right belowSearch'>
                <li className='inline-block bg-black/50 p-2'>
                    <a href="/shelves"><img src="/img/icons/shelves.png" className="inline w-6" alt="" />Or see the <span className='font-bold underline'>3D shelves!</span></a>
                </li>
            </ul>
        </div>
        <div className='w-[100%] p-5 ml-auto mr-auto max-w-6xl text-white'>
            <div className='flex flex-col lg:flex-row items-start gap-2'>
              <div className='flex-1'>
                <div id="home-info" className='bg-black/50 p-5 w-[100%]'>
                  <HomeBlurb />
                </div>
                <div id="latest-list" className='bg-black/50 p-5 w-[100%] mt-3'>
                  <h2 className='text-[22px] -mt-2 text-center block font-bold mb-5'>Latest Additions</h2>
                  <table className='table-auto md:table-fixed w-full'>
                    <thead>
                      <tr>
                        <th className='text-left'>Title</th>
                        <th className='text-right'>Date Added</th>
                      </tr>
                    </thead>
                    <tbody>
                      {latestGames && latestGames.length > 0 && (
                        latestGames.map((game: any) => (
                          <tr key={game.id}>
                            <td className='p-1'>
                              <a href={"/game/"+game.slug}>
                                <img className='inline-block mr-2' src={"/scans/"+game.slug+"/front.webp"} width={24} alt={game.title} />
                                <span className='underline'>{game.title}</span>
                              </a>
                            </td>
                            <td className='p-1 text-right'>{new Date(game.created_at).toLocaleDateString()}</td>
                          </tr>
                        ))
                      )}
                    </tbody>
                    
                  </table>
                </div>
              </div>
              <div id="random-game" className='overflow-hidden relative w-full lg:w-80 lg:max-w-80 lg:flex-shrink-0 bg-black/50 p-5 h-150 max-w-6xl'>
                <h2 className='text-[22px] text-center font-bold'>Box Of The Moment!</h2>
                {botd && <SingleGame ga={botd} zd={-6} showTitle={true} showFooter={true} />}
              </div>
            </div>
        </div>
      </div>
    )
}