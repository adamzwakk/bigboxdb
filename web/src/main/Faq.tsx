import Header from '@/partials/Header';
import '../globals.css'
import './main.scss'
import { useEffect, useState } from 'react';

export default function Faq()
{
    const [stats, setStats] = useState<Array<{name:string,count:number}>>()

    useEffect(() => {
        fetch('/api/variants/typecount')
            .then(res => res.json())
            .then((data) => {
                setStats(data)
            })
            .catch(console.error)
    },[])

    return(
        <div className='relative z-4 text-white'>
            <title>FAQ | BigBoxDB</title>
            <Header />
            <div id="faq" className='bg-black/50 w-[95%] p-5 mt-10 ml-auto mr-auto max-w-4xl text-white'>
                <div id="faq-content">
                    <h3 className='mb-3 font-bold'>What is this?</h3>
                    <p className='mb-3'>It&apos;s a website dedicated to trying to organize and categorize very old computer games when physical media still mattered to everyone. I found that the internet is pretty lacking in databasing these properly, so I&apos;m trying to do what I can. I orignally got really inspired after discovering <a href="https://bigboxcollection.com/" className='underline' target='_blank'>BigBoxCollection</a> and wanted to do my own spin on the idea.</p>
                    <h3 className='mb-3 font-bold'>Why are you doing this?</h3>
                    <p className='mb-3'>My current solution of writing what I had down and taking pictures wasn&apos;t super thrilling. So I thought I might as well give back to the internet at the same time.</p>
                    <h3 className='mb-3 font-bold'>How are you doing this?</h3>
                    <p>I&apos;m scanning all of my personal collection. Boxes are scanned at 800 or 1200 DPI. I&apos;ve found that &gt;1200 DPI for box scans seems to be the sweet spot for box art while still being way over the top. The Epson does 4800 DPI and it&apos;s just not worth the space and boxes were definitely not printed that small to begin with. <br/><br/>My current setup is:</p>
                    <ul className='mb-3'>
                        <li>Epson Perfection V600 Flatbed Scanner</li>
                        <li className='line-through'>Epson Perfection V39 Flatbed Scanner (retired)</li>
                        <li><a href="https://www.hamrick.com/">VueScan</a></li>
                    </ul>
                    <h3 className='mb-3 font-bold'>How many games are here?</h3>
                    <ul className='mb-3'>
                        {stats && stats.map((s, index) => <li key={index} dangerouslySetInnerHTML={{__html: s['name']+': '+s['count']}}></li> )}
                    </ul>
                    <h3 className='mb-3 font-bold'>How did you make this?</h3>
                    <p className='mb-3'>The 3D Boxes and Shelves I&apos;m using Three.JS/react-three-fiber along with a custom conversion script in python to generate glb files with ktx2 textures, a low def one for the shelf, and a higher def one for when you pick up the box for a closer look. I still retain my old CSS-only version <a href="https://www.bigboxdb.com/shelves/css" className="underline">here</a> for my own amusement/prove that I could do it, though it always used the threejs boxes for the closer look. Kind of threw myself into a bunch of new tech at once to force myself to learn it. Maybe I&apos;ll do a writeup later more about the tooling if there&apos;s enough interest.</p>
                    <h3 className='mb-3 font-bold'>Are you accepting community scans? Can I help?</h3>
                    <p className='mb-3'>Of course! I want to open this up to more contributions once I have this a bit more nailed down :D If you have something to contribute or have general feedback, email me at <a className='underline' href="mailto:adam@adamzwakk.com">adam@adamzwakk.com</a> or check out our <a className='underline' href="https://discord.gg/CWXbgxCeW5">Discord</a>.</p>
                    <h3 className='mb-3 font-bold'>Wait! Small boxes too?! That&apos;s not the name of the site!</h3>
                    <p className='mb-3'>I feel like there is more big boxes than small boxes, and I assume no one has any interest in a &quot;smallboxdb&quot; so here we are, deal with it.</p>
                    <h3 className='mb-3 font-bold'>Can I use your scans/work?</h3>
                    <p className='mb-3'>Totally cool with that! Just make sure to give me credit/thanks in some way!</p>
                    <h3 className='mb-3 font-bold'>General legal stuff</h3>
                    <p className='mb-3'>I acknowledge that even though I physically own these games, I don&apos;t have the rights to distribute them. I won&apos;t ever post files/actual software, this site is purely for information and sharing artwork/documentation. If there is something here that you own and would like taken down feel free to email me and I will gladly discuss and/or take down the content.</p>
                </div>
            </div>
        </div>
    )
}