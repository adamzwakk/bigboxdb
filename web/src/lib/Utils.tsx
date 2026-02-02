import { useEffect, useState } from "react";

export function IsTouchDevice(){
  if (typeof window !== "undefined") {
      return (('ontouchstart' in window) ||
          (navigator.maxTouchPoints > 0) ||
          // @ts-ignore
          (navigator.msMaxTouchPoints > 0));
    }
}

export function useWindowSize() {
    const [windowSize, setWindowSize] = useState({
      width: 0,
      height: 0,
    });
  
    useEffect(() => {
      function handleResize() {
        setWindowSize({
          width: window.innerWidth,
          height: window.innerHeight,
        });
      }
      
      window.addEventListener("resize", handleResize);
       
      handleResize();
      
      return () => window.removeEventListener("resize", handleResize);
    }, []); 
    return windowSize;
  }