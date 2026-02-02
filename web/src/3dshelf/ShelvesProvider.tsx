import { IsTouchDevice } from '@/lib/Utils';
import type { Game3D } from '@/lib/types';
import { createContext, useContext, useEffect, useState } from 'react';

const ShelvesContext = createContext<any>(null);

export function useShelvesData() {
  return useContext(ShelvesContext);
}

export default function ShelvesProvider({ 
  children 
}: { 
  children: React.ReactNode 
}) {
  const [allGames, setAllGames] = useState<Array<Game3D>>([]);
  const [developers,setDevelopers] = useState<any>()
  const [publishers,setPublishers] = useState<any>()
  const [shelfLength, setShelfLength] = useState(IsTouchDevice() ? 40 : 100)

  useEffect(() => {
    fetch('/api/variants/all')
      .then(res => res.json())
      .then((data: Game3D[]) => {
          setAllGames(data)
      })
      .catch(console.error)

    let devs:any = []
    fetch('/api/developers/all')
      .then((res) => res.json())
      .then((s:any) => {
          s.forEach((d: { slug: any; name: any; softwareCount:number }) => {
              devs.push({slug:d.slug,name:d.name,count:d.softwareCount})
          });
          setDevelopers(devs)
      })

    let pubs:any = []
    fetch('/api/publishers/all')
      .then((res) => res.json())
      .then((s:any) => {
          s.forEach((d: { slug: any; name: any; softwareCount:number }) => {
              pubs.push({slug:d.slug,name:d.name,count:d.softwareCount})
          });
          setPublishers(pubs)
      })
  }, []); 

  const contextValue = {
    allGames,
    developers,
    publishers,
    shelfLength,
    setShelfLength
  }

  return (
    <ShelvesContext.Provider value={contextValue}>
      {children}
    </ShelvesContext.Provider>
  );
}