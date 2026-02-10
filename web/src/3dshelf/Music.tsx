import { useRef } from "react";

export default function Music() {
    const thisAudio = useRef<HTMLAudioElement>(null);

    const playMusic = function()
    {
        if(thisAudio.current)
        {
            thisAudio.current.muted = !thisAudio.current.muted;
            if(thisAudio.current.paused)
            {
                thisAudio.current.play();
            }
        }
    }
    
    return (
        <div id="music" className="z-10">
            <a href="#" onClick={playMusic}><img src="/img/3dshelf/mute.png" title="Toggle Music" width={20} height={20} />&nbsp;</a>
            <audio ref={thisAudio} loop muted>
                <source src="/music/keen4e.ogg" type="audio/ogg" />
            </audio> 
        </div>
    )
}