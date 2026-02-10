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
          s.forEach((d: { id:number, slug: string; name: string; variant_count:number }) => {
              devs.push({id:d.id, slug:d.slug,name:d.name,count:d.variant_count})
          });
          setDevelopers(devs)
      })

    let pubs:any = []
    fetch('/api/publishers/all')
      .then((res) => res.json())
      .then((s:any) => {
          s.forEach((d: { id:number, slug: string; name: string; variant_count:number }) => {
              pubs.push({id:d.id, slug:d.slug,name:d.name,count:d.variant_count})
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