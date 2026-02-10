import {useStore} from "@/lib/Store";
import { useNavigate, useParams } from "react-router";

export default function BackBlur() {

    const { isDragging } = useStore();
    const navigate = useNavigate();
    const params = useParams()

    const handleClick = function(e:any){
        e.stopPropagation()
        if(isDragging) { return; }
        navigate("/shelves");
    }

    return(
        params.variantId && 
        <mesh position-z={9} onPointerUp={(e) => {e.stopPropagation()}} onPointerEnter={(e) => {e.stopPropagation()}} onClick={handleClick}>
            <planeGeometry args={[1000,1000]} />
            <meshBasicMaterial color={0x000000} transparent={true} opacity={0.7} />
        </mesh>
    )
}