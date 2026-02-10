import {useLoader, useThree} from '@react-three/fiber'
import React, {Suspense, useEffect, useMemo, useRef, useState} from "react";
import {find} from "lodash";
import { useGesture } from '@use-gesture/react'
import gsap from 'gsap';
import * as THREE from 'three';
import { GLTFLoader } from 'three/addons/loaders/GLTFLoader.js';
import { KTX2Loader } from 'three/addons/loaders/KTX2Loader.js';
// import { sendGTMEvent } from '@next/third-parties/google'

import {useStore} from "@/lib/Store";
import {IsTouchDevice} from "@/lib/Utils";
import type { Game3D } from '@/lib/types';
import { AllGatefoldTypes, BigBoxTypes, BoxShelfDirection, VerticalGatefoldTypes } from '@/lib/enums';
import { useNavigate, useParams } from 'react-router';

type BigBoxProps = {
    position: {x:number, y:number, z:number},
    g: Game3D,
    onShelf:boolean
}

// Singleton KTX2 loader - shared across all models
let ktx2LoaderInstance: KTX2Loader | null = null;

function getKTX2Loader(gl: THREE.WebGLRenderer): KTX2Loader {
    if (!ktx2LoaderInstance) {
        ktx2LoaderInstance = new KTX2Loader();
        ktx2LoaderInstance.setTranscoderPath('/basis/');
        ktx2LoaderInstance.detectSupport(gl);
    }
    return ktx2LoaderInstance;
}

// Hook to detect if object is in camera frustum
function useInView(ref: React.RefObject<THREE.Group | null>, margin = 1.5) {
    const [isInView, setIsInView] = useState(false);
    const { camera } = useThree();
    const frustum = useMemo(() => new THREE.Frustum(), []);
    const matrix = useMemo(() => new THREE.Matrix4(), []);

    useEffect(() => {
        const checkVisibility = () => {
            if (!ref.current) return;

            matrix.multiplyMatrices(camera.projectionMatrix, camera.matrixWorldInverse);
            frustum.setFromProjectionMatrix(matrix);

            // Create a bounding sphere around the object's position with margin
            const sphere = new THREE.Sphere(ref.current.position, margin);
            const inView = frustum.intersectsSphere(sphere);
            
            setIsInView(inView);
        };

        // Check immediately
        checkVisibility();

        // Check periodically (you can adjust frequency)
        const interval = setInterval(checkVisibility, 200);
        return () => clearInterval(interval);
    }, [camera, frustum, matrix, ref, margin]);

    return isInView;
}

// Lazy-loaded model component with KTX2 support
function LazyModel({ g, gatefoldRef, useHighQuality, onShelf }: { g: Game3D, gatefoldRef: React.RefObject<THREE.Object3D | null>, useHighQuality: boolean, onShelf: boolean }) {
    const { gl } = useThree();
    const modelRef = useRef<THREE.Group>(null);
    const modelPath = useMemo(() => {
        if (!g.textureFileName) return '';
        
        if (!useHighQuality) {
            return g.textureFileName.replace('.glb','-low.glb');
        } else {
            return g.textureFileName;
        }
    }, [g.textureFileName, useHighQuality]);
    
    // Get shared KTX2 loader instance
    const ktx2Loader = useMemo(() => getKTX2Loader(gl), [gl]);
    
    // Load GLTF with KTX2 loader configured
    const gltf = useLoader(
        GLTFLoader,
        modelPath,
        (loader) => {
            loader.setKTX2Loader(ktx2Loader);
        }
    );

    const boxObj = gltf.scene;
    useEffect(() => {
        if (modelRef.current && onShelf) {
            if(g.dir == BoxShelfDirection.front)
            {
                modelRef.current.position.set(0, 0, 2);
            }
            else
            {
                modelRef.current.position.set(-2, 0, 0);
            }

            const delay = Math.random() * 400;
            setTimeout(() => {
                if (!modelRef.current) return;
                gsap.to(modelRef.current.position, {
                    x: 0,
                    y: 0,
                    z: 0,
                    duration: 0.6,
                    ease: "back.out(1.7)"
                });
            }, delay);
        }
    }, []);
    
    useEffect(() => {
        if(boxObj) {
            boxObj.traverse((child) => {
                if(child instanceof THREE.Mesh) {
                    
                    if(child.material) {
                        const materials = Array.isArray(child.material) ? child.material : [child.material];
                        materials.forEach(mat => {
                            mat.side = THREE.FrontSide
                            
                            mat.depthTest = true;
                            mat.depthWrite = true;
                        });
                    }
                }
                if(child.name === 'Gatefold' && child instanceof THREE.Mesh) {
                    gatefoldRef.current = child;
                    
                    child.geometry.computeBoundingBox();
                    let bounds = child.geometry.boundingBox;
                    
                    if(!bounds) return;
                    
                    if(VerticalGatefoldTypes.has(g.box_type)) {
                        // VERTICAL GATEFOLD - hinge at top, opens upward
                        const topEdge = bounds.max.y;
                        const translateAmountY = -topEdge;
                        const zCenter = (bounds.min.z + bounds.max.z) / 2;
                        
                        child.geometry.translate(0, translateAmountY, -zCenter);
                        child.geometry.computeBoundingBox();
                        child.position.set(0, g.h!/2, g.d!/2);
                        
                    } else if(g.box_type === BigBoxTypes.Big_Box_With_Back_Gatefold) {
                        // BACK GATEFOLD - hinge on left, but at back face
                        const leftEdge = bounds.min.x;
                        const translateAmount = -leftEdge;
                        const zCenter = (bounds.min.z + bounds.max.z) / 2;
                        
                        child.geometry.translate(translateAmount, 0, -zCenter);
                        child.geometry.computeBoundingBox();
                        
                        // Position at left edge and BACK face (negative Z)
                        child.position.set(-g.w!/2, 0, -g.d!/2);
                                                
                    } else {
                        // STANDARD HORIZONTAL GATEFOLD
                        const leftEdge = bounds.min.x;
                        const translateAmount = -leftEdge;
                        const zCenter = (bounds.min.z + bounds.max.z) / 2;
                        
                        child.geometry.translate(translateAmount, 0, -zCenter);
                        child.geometry.computeBoundingBox();
                        child.position.set(-g.w!/2, 0, g.d!/2);
                    }
                }
                
                // Handle transparency for gatefold materials
                if(g.gatefold_transparent && child instanceof THREE.Mesh && child.material) {
                    if(Array.isArray(child.material)) {
                        child.material.forEach(mat => {
                            if(mat instanceof THREE.MeshStandardMaterial) {
                                mat.transparent = true;
                            }
                        });
                    } else if(child.material instanceof THREE.MeshStandardMaterial) {
                        child.material.transparent = true;
                    }
                }
            });
        }
    }, [boxObj]);

    return <primitive ref={modelRef} object={boxObj} scale={1} />;
}

// Placeholder component while loading
function BoxPlaceholder({ g }: { g: Game3D }) {
    return (
        <mesh>
            <boxGeometry args={[g.w, g.h, g.d]} />
            <meshBasicMaterial color="#333333" opacity={0.3} transparent />
        </mesh>
    );
}

function BigBox({position, g, onShelf}:BigBoxProps)
{
    const navigate = useNavigate();
    const params = useParams()

    const groupRef = useRef<THREE.Group>(null);
    const rotateXTo = useRef<gsap.QuickToFunc | null>(null);
    const rotateYTo = useRef<gsap.QuickToFunc | null>(null);

    const gatefoldRef = useRef<THREE.Object3D | null>(null);
    const {camera, size, viewport} = useThree()
    const aspect = size.width / viewport.width;
    const [hovering, setHovering] = useState(false)
    const [active, setActive] = useState(false)
    const [gatefoldOpen, setGatefoldOpen] = useState(false)
    const [useHighQuality, setUseHighQuality] = useState(false);
    
    const [previousMouseDelta, setPreviousMouseDelta] = useState([0,0])
    const { shouldHover, setIsDragging } = useStore();
    const zDelta = IsTouchDevice() ? 12 : 10

    // Check if this box is in view
    const isInView = useInView(groupRef, 0.01); // 5 unit margin
    const [shouldLoad, setShouldLoad] = useState(false);

    // Stagger loading to avoid loading many boxes at once
    useEffect(() => {
        if (isInView && !shouldLoad) {
            // Random delay between 0-200ms to stagger loads
            const delay = Math.random() * 200;
            const timeout = setTimeout(() => {
                setShouldLoad(true);
            }, delay);
            return () => clearTimeout(timeout);
        }
    }, [isInView, shouldLoad]);

    const bind = useGesture({
        onDragStart: () => {
            if(active)
            {
                setIsDragging(true)
            }
        },
        onDrag: ({ delta: [x, y], timeStamp, event, tap, movement: [mx, my], ctrlKey }) => {
            event.stopPropagation();

            if(tap || (Math.abs(mx) < 2 && Math.abs(my) < 2)) {
                return timeStamp;
            }
            
            if(active && groupRef.current)
            {  
                if(ctrlKey)
                {
                    const moveSensitivity = 0.02; // Adjust this to change movement speed
                    
                    groupRef.current.position.x += x * moveSensitivity;
                    groupRef.current.position.y -= y * moveSensitivity; // Negative because screen Y is inverted
                }
                else
                {
                    x = previousMouseDelta[0] + x;
                    y = previousMouseDelta[1] + y;
                    
                    rotateXTo.current?.(y / aspect);
                    rotateYTo.current?.(x / aspect);
                    
                    setPreviousMouseDelta([x,y]);
                }
            }
            return timeStamp;
        },
        onDragEnd: () => {
            if(active)
            {
                setTimeout(() => {
                    setIsDragging(false);
                }, 150);
            }
        },
        onWheel: (e:any) => {
            if(active && e.wheeling && groupRef.current)
            {
                let direction = groupRef.current.position.z + (-e.direction[1]*2);
                
                if(direction <= 6) {
                    return
                }
                gsap.to(groupRef.current.position, {
                    z: direction,
                    duration: 0.3,
                    ease: "power2.out"
                });
            }
        }
    }, {
        drag: {
            filterTaps: true
        }
    });

    useEffect(() => {
        if (groupRef.current) {
            rotateXTo.current = gsap.quickTo(groupRef.current.rotation, "x", {
                duration: 0.5,
                ease: "power2.out"
            });
            rotateYTo.current = gsap.quickTo(groupRef.current.rotation, "y", {
                duration: 0.5,
                ease: "power2.out"
            });
        }
        if(!onShelf)
        {
            setActive(true)
            setUseHighQuality(true)
        }
    },[]); 

    useEffect(() => {
        if(onShelf){
            if(params.variantId && parseInt(params.variantId) == g.id && !active)
            {
                activateGame(null)
                setActive(true)
            }
            else if(!params.variantId)
            {
                setActive(false)
            }
        }
    },[params])

    useEffect(() => {
        if(find(shouldHover,{id:g.id}))
        {
            hover(null)
        }
        else
        {
            unhover(null)
        }
    }, [shouldHover]);

    useEffect(() => {
        if(!active && groupRef.current){
            gsap.to(groupRef.current.position, {
                z: hovering ? position.z + 2 : position.z,
                duration: 0.15,
                ease: "power2.out"
            });
        }
    },[hovering])

    useEffect(() => {
        if(!groupRef.current) return;

        if(active)
        {
            gsap.to(groupRef.current.position, {
                x: camera.position.x,
                y: camera.position.y,
                z: onShelf ? camera.position.z - zDelta : position.z,
                duration: 0.15,
                ease: "power2.out"
            });
            gsap.to(groupRef.current.rotation, {
                x: 0,
                y: 0.4,
                z: 0,
                duration: 0.15,
                ease: "power2.out"
            });
        }
        else
        {
            setPreviousMouseDelta([0,0]);
            gsap.to(groupRef.current.position, {
                x: position.x,
                y: position.y,
                z: position.z,
                duration: 0.15,
                ease: "power2.out"
            });
            gsap.to(groupRef.current.rotation, {
                x: 0,
                y: (g.dir == BoxShelfDirection.left ? Math.PI / 2 : 0),
                z: 0,
                duration: 0.15,
                ease: "power2.out"
            });
        }
    },[active])

    useEffect(() => {
        if(!gatefoldRef.current)
        {
            return
        }
        const gatefold = gatefoldRef.current;
        if(gatefoldOpen)
        {
            if(VerticalGatefoldTypes.has(g.box_type)) {
                gsap.to(gatefold.rotation, {
                    x: -Math.PI,
                    duration: 0.5,
                    ease: "power2.inOut"
                });
                // gsap.to(group!.position, {
                //     y: group!.position.y-(g.h/2),
                //     duration: 0.5,
                //     ease: "power2.inOut"
                // });
            } else if(g.box_type === BigBoxTypes.Big_Box_With_Back_Gatefold) {
                gsap.to(gatefold.rotation, {
                    y: Math.PI,
                    duration: 0.5,
                    ease: "power2.inOut"
                });
            } else {
                gsap.to(gatefold.rotation, {
                    y: -Math.PI,
                    duration: 0.5,
                    ease: "power2.inOut"
                });
                // gsap.to(group!.position, {
                //     x: group!.position.x+(g.w/2),
                //     duration: 0.5,
                //     ease: "power2.inOut"
                // });
            }
        }
        else
        {
            gsap.to(gatefold.rotation, {
                x: 0,
                y: 0,
                duration: 0.5,
                ease: "power2.inOut"
            });
            // if(VerticalGatefoldTypes.has(g.box_type)) 
            // {
            //     gsap.to(group!.position, {
            //         y: group!.position.y-(g.h/2),
            //         duration: 0.5,
            //         ease: "power2.inOut"
            //     });
            // }
            // else
            // {
            //     gsap.to(group!.position, {
            //         x: group!.position.x-(g.w/2),
            //         duration: 0.5,
            //         ease: "power2.inOut"
            //     });
            // }
        }
    },[gatefoldOpen])

    const hover = function(e:any)
    {
        if(e !== null) { e.stopPropagation() }
        if(active){
            document.body.style.cursor = 'move';
            return;
        }
        document.body.style.cursor = 'pointer';
        setHovering(true)
    }

    const unhover = function(e:any)
    {
        if(e !== null) { e.stopPropagation() }
        document.body.style.cursor = 'default';
        if(active){return;}
        setHovering(false)
    }

    const activateGame = function(e:any){
        if(e !== null) { e.stopPropagation() }
        if(active)
        {
            e?.stopPropagation()
            return
        }

        setUseHighQuality(true);
        setHovering(false)

        if(onShelf)
        {
            navigate("/shelves/game/"+g.slug);
            // window.history.replaceState(null, '', '/shelves/game/'+g.slug)
            // document.title = 'BigBoxDB | '+g.title
            // sendGTMEvent({ event: 'page_view', pagePath: '/shelves/game/'+g.slug })
        }
    }

    const togGatefold = function(forceClose = false)
    {
        if(!active || !AllGatefoldTypes.has(g.box_type))
        {
            return
        }
        if(!gatefoldOpen)
        {
            setGatefoldOpen(true);
        }
        else if(forceClose || gatefoldOpen)
        {
            setGatefoldOpen(false);
        }
    }

    return ( 
        <group 
            ref={groupRef}
            position={[position.x, position.y, position.z]}
            rotation={[0, (g.dir == BoxShelfDirection.left ? Math.PI / 2 : 0), 0]}
            {...bind()}
            onPointerOver={!onShelf || !params.variantId ? hover : undefined}
            onPointerOut={unhover}
            onClick={activateGame}
            onDoubleClick={togGatefold}
            castShadow={false}
        >
            {shouldLoad ? (
                <Suspense fallback={<BoxPlaceholder g={g} />}>
                    <LazyModel g={g} gatefoldRef={gatefoldRef} useHighQuality={useHighQuality} onShelf={onShelf} />
                </Suspense>
            ) : (
                <BoxPlaceholder g={g} />
            )}
        </group>
    )
}

export default BigBox;