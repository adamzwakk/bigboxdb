'use client'

import type { StatsProps } from "@/lib/types"
import { useEffect, useState } from "react"
import { findIndex } from "lodash"

export default function HomeBlurb()
{
    const [stats,setStats] = useState<Array<StatsProps>>()

    useEffect(() => {
        const res = fetch('/api/stats')
            .then((res) => res.json())
            .then((s:Array<StatsProps>) => {
                setStats(s)
            })
    },[])

    return(
        <div id="home-info" className='bg-black/50 p-5 mt-5 w-[100%] ml-auto mr-auto max-w-2xl'>
            <h2 className='text-[22px] -mt-2 text-center block font-bold mb-5'>What? No one else remembers the best kind of physical media?</h2>
            <p className='mb-5'>I have been scanning my whole collection of big box (and to a lesser extent small box) PC games in an effort to make the box art more available on the internet. Hopefully this fills some gaps, fulfills nostalgia, and helps anyone looking for more info about this old type of physical media.</p>
            <p className='mb-5'>Plus I mean don't you just <a href="/img/lovegraphics.jpg" className='underline' target="_blank">love the graphics on this box?</a></p>
            {(stats && stats.length && <p className='mb-5'>So far we have <span dangerouslySetInnerHTML={{__html: stats[findIndex(stats, function(o) { return o.name == 'Total Boxes'; })].count}}></span> total boxes scanned!</p>)}
        </div>
    )
}