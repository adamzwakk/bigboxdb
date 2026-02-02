"use client"

import { DoubleSide } from "three";

function Wall()
{
    return(
        <mesh position-z={-23}>
            <planeGeometry args={[1000,1000]} />
            <meshBasicMaterial color={0xfcfced} side={DoubleSide} />
        </mesh>
    )
}

export default Wall;
