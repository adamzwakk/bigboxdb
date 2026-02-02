import {useStore} from "@/lib/Store";
import {isEmpty} from "lodash";

export default function BackBlur() {

    const { activeGame,setActiveGame, setControlsEnable, isDragging } = useStore();

    const handleClick = function(e:any){
        e.stopPropagation()
        if(isDragging) { return; }
        if(!isEmpty(activeGame))
        {
            window.history.replaceState(null, '', '/shelves')
            document.title = 'BigBoxDB | 3D Shelves'
            
            setControlsEnable(true)
            setActiveGame(null)
        }
    }

    return(
        !isEmpty(activeGame) && 
        <mesh position-z={9} onPointerUp={(e) => {e.stopPropagation()}} onPointerEnter={(e) => {e.stopPropagation()}} onClick={handleClick}>
            <planeGeometry args={[1000,1000]} />
            <meshBasicMaterial color={0x000000} transparent={true} opacity={0.7} />
        </mesh>
    )
}