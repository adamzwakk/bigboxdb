import {has,forEach, isNumber} from "lodash"
import { ImageOrder } from "./enums";
import { BoxSides } from "./types";

export function config()
{
	return {
		'scan_path':'/scans',
		'2ddpi': 30,
		'shelfLength': 100,
		'default3dtexture':'/img/default_box_texture.jpg'
	}
}

