import { BoxGeometry, BufferGeometry } from "three"
import { BigBoxTypes, BoxShelfDirection } from "./enums"

export type ShelvesProps = {
    shelves:Array<ShelfProps>
}

export type ShelfProps = {
    games:Array<Game3D>,
    shelfY:number,
    width:number
}

export type BoxProps = {
	g:Game3D,
    active?:boolean,
	position:{
        x:number,
        y:number
    }
}

export type Box3DProps = BoxProps & {
	position: {x:number, y:number, z:number}
}

export type Box2DProps = BoxProps & {
    g: Game
}

export type SearchIndex = {
	id:number,
	name?:string,
	slug?:string,
	x?:number,
	y?:number
}

type BoxLinks = {
	steam: string,
	gog: string
}

export type BoxSides = {
	front: string,
	left: string,
	right: string,
	back: string,
	top: string,
	bottom: string,
	gatefold0?: string,
	gatefold1?: string,
	gatefold2?: string,
	gatefold3?: string
}

export type Game = {
	// initial values
	id: number,
	title: string,
	slug: string,
	images?: BoxSides
	box_type: BigBoxTypes,
	year:number,
	platform:string,
	variant:string,
	
	steam_link?:string
	gog_link?:string,
	other_link?:string,
	series?: string,

	gatefold_transparent: boolean,
	dir: BoxShelfDirection,
	
	w: number,
	d: number,
	h: number,

	scan_notes?: string
	contributor?: string
	description?: string
	worth_front_view?: boolean
	publishers?:Array<any>
	developers?:Array<any>
}

export type Game3D = Game &
{
	sd: number,
	shelfX?: number,
	shelfY?: number,
    shelfZ?: number,
	textureFileName?: string,
    boxGeo?: BufferGeometry|BoxGeometry,
}

export type StatsProps = {
    name:string,
    count:number
	id?: number
}

export type GameInfoProps = {
	title: string,
	year: number,
	width: number,
	height: number,
	depth: number,
	images: BoxSides,
	platform: string,
	variant: string,
	scan_notes: string,
	type:BigBoxTypes,
	contributed_by:string,

	links: BoxLinks
}

export type BigBoxTypeOrAbandonware = BigBoxTypes | -1;