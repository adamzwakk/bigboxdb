import {create} from "zustand";
import type { Game3D, SearchIndex, ShelfProps } from "@/lib/types";

interface BigBoxDBState {
    goToSearchedGame:{variant_id:number,slug:string}|null
    shouldHover:Array<SearchIndex>|null
    controlsEnable:boolean
    toggleGatefold:boolean
    shelfNum:number,
    isDragging: boolean,
    activeShelves:Array<ShelfProps>,
    stagedOptions:{shelfLength:number,boxTypes:Array<number>,dev:string|null,pub:string|null}
    setGoToSearchedGame:(g:{variant_id:number,slug:string}|null) => void
    setControlsEnable:(e:boolean) => void
    setShouldHover: (g:Array<SearchIndex>) => void
    setShelfNum:(n:number) => void
    setIsDragging:(e:boolean) => void
    setActiveShelves: (games:Array<ShelfProps>) => void
    setStagedOptions:(options:any) => void
}

export const useStore = create<BigBoxDBState>((set) => ({
    controlsEnable:true,
    shouldHover:null,
    goToSearchedGame:null,
    toggleGatefold:false,
    shelfNum:0,
    isDragging: false,
    activeShelves: [],
    stagedOptions:{shelfLength:100, boxTypes:[0],dev:null,pub:null},

    setShelfNum: (e) => {
        set(() => ({
            shelfNum: e
        }));
    },

    setGoToSearchedGame: (e: {variant_id:number,slug:string}|null) => {
        set(() => ({
            goToSearchedGame: e
        }));
    },

    setControlsEnable: (e) => {
        set(() => ({
            controlsEnable: e
        }));
    },

    setShouldHover: (g: Array<SearchIndex>) => {
        set(() => ({
            shouldHover: g
        }));
    },

    setIsDragging: (dragging: boolean) => set({ isDragging: dragging }),

    setActiveShelves: (e) => {
        set(() => ({
            activeShelves: e
        }));
    },

    setStagedOptions: (e) => {
        set(() => ({
            stagedOptions: e
        }));
    },
}));
