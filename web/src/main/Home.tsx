import '../globals.css'
import './main.scss'
import HomeBlurb from './partials/HomeBlurb'

export default function Home() {
    return(
        <div id="home" className='relative'>
        <div id="main-header" className="ml-auto mr-auto w-full sm:w-lg p-5 z-4 text-white pt-[3vh]">
            <a href="/">
                <h1 className='text-[50px] text-center block font-bold'><img className='mainLogo inline w-[50px] mr-2 mb-2' src="/img/logo_filled.png" alt="Logo"/> BigBoxDB</h1>
                <h2 className='text-[18px] -mt-2 text-center block font-bold'>an elegant wrapping from a more civilized age</h2>
            </a>
            {/* <Search onShelf={false} /> */}
            <ul className='text-right belowSearch -mt-3'>
                <li className='inline-block bg-black/50 p-2'>
                    <a href="/shelves"><img src="/img/icons/shelves.png" className="inline w-6" alt="" />Or see the <span className='font-bold underline'>3D shelves!</span></a>
                </li>
            </ul>
        </div>
        <div className='w-[100%] p-5 ml-auto mr-auto max-w-4xl text-white'>
            <HomeBlurb />

            <div id="random-game" className='bg-black/50 p-5 mt-5 w-[95%] ml-auto mr-auto max-w-2xl'>
              <h2 className='text-[22px] text-center font-bold'>BOTD!</h2>
              {/* <SingleGame slug={false} zd={0} showFooter={true} /> */}
            </div>
        </div>
    </div>
    )
}