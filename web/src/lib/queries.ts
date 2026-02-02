import path from 'path';

import Redis from "ioredis";
import { findIndex, forEach, sortBy, filter } from "lodash";
import { StatsProps, Game, Game3D, MariaGame } from "./types";
import { AllGatefoldTypes, BigBoxTypes, BoxShelfDirection } from "./enums";
import slugify from "slugify";
import db from "@/models";
import { col, fn, literal } from 'sequelize';

let redis:Redis|null = null
if (process.env.BUILD === undefined) redis = new Redis(6379, 'redis');

export async function queryGames(slug:null|string, fields:any = []) : Promise<Array<Game3D>>
{
	let cleaned:Array<Game3D> = []
	const env = process.env.NODE_ENV
	let cache
	let g:Game3D
	if(redis)
	{
		cache = await redis.get("all_games_frontend");
	}
	if(env == "production" && cache)
	{
		cleaned = JSON.parse(cache)
	}
	else
	{
		const games = await db.Software.findAll({
			include: [
				{
					model: db.Platform,
					as: 'platform',
					attributes:['name']
				},
				{
					model: db.Developer,
					as: 'developers',
					attributes:['slug']
				},
				{
					model: db.Publisher,
					as: 'publishers',
					attributes:['slug']
				},
				{
					model: db.User,
					as: 'contributor',
					attributes:['name'],
					
				},
			],
			order:[
				[literal('COALESCE(series, title)'), 'DESC'],
			]
		})

		let sorted = sortBy(games, function(g){ 
			if(g.series)
			{
				return g.series.toLocaleLowerCase()
			}

			return g.title.toLocaleLowerCase()
		})
			
		let lastGame:Game3D
		for (const ga of sorted) {
			g = cleanupGame(ga as unknown as MariaGame)

			lastGame = g
			cleaned.push(g)
		}

		if(redis)
		{
			await redis.set("all_games_frontend", JSON.stringify(cleaned), "EX", 3600);
		}
	}

	if(slug !== null)
	{
		cleaned = filter(cleaned,{slug:slug})
	}

	return cleaned
}

export async function queryGame(criteria:any, fields:any = []) : Promise<Game3D>
{
	const env = process.env.NODE_ENV
	let cacheKey = slugify(JSON.stringify(criteria)+'_'+JSON.stringify(fields)).replace(/["\:]/g,'')

	let cache
	if(redis)
	{
		cache = await redis.get("game_info_"+cacheKey);
	}
	if(env == "production" && cache)
	{
		return JSON.parse(cache)
	}
	else
	{
		const ga = await db.Software.findOne({
			where:criteria,
			include: [
				{
					model: db.Platform,
					as: 'platform',
					attributes:['name'],
				},
				{
					model: db.User,
					as: 'contributor',
					attributes:['name'],
				},
				{
					model: db.Developer,
					as: 'developers',
					attributes: ['name'],
					through: { attributes: [] },
				},
				{
					model: db.Publisher,
					as: 'publishers',
					attributes: ['name'],
					through: { attributes: [] },
				},
			],
			attributes: fields.length > 0 ? fields : undefined
		})

		let g = cleanupGame(ga as unknown as MariaGame,true)

		if(redis)
		{
			await redis.set("game_info_"+cacheKey, JSON.stringify(g), "EX", 3600);
		}

		return g
	}
}

function cleanupGame(ga:MariaGame,morePlz:boolean = false)
{
	let g:Game3D = {
		id: ga.id!,
		title: ga.title,
		slug: ga.slug,
		year: ga.year,
		platform: ga.platform?.name as string,
		variant: ga.variant,
		w: ga.width,
		h: ga.height,
		d: ga.depth,
		dir: BoxShelfDirection.left,
		worth_front_view:ga.worth_front_view,
		box_type: ga.box_type_id,
		gatefold_transparent: false,
		textureFileName: '/'+path.join('scans', ga.slug, 'box.glb'),
		sd:0,
		shelfX:0,
		shelfY:0,
		shelfZ:0,
		developers:ga.developers,
		publishers:ga.publishers
	}

	if(morePlz)
	{
		if(ga.description)
		{
			g.description = ga.description
		}
		if(ga.scan_notes)
		{
			g.scan_notes = ga.scan_notes
		}
	}

	if(ga.box_type_id == 2)
	{
		// eidos trapezoid
		g.w = 10
		g.h = 10
		g.d = 2
		g.dir = BoxShelfDirection.front
	}

	if(ga.series)
	{
		g.series = ga.series
	}

	if(ga.steam_link)
	{
		g.steam_link = ga.steam_link
	}

	if(ga.gog_link)
	{
		g.gog_link = ga.gog_link
	}

	if(ga.contributor)
	{
		g.contributor = ga.contributor.name as string
	}

	if(AllGatefoldTypes.has(ga.box_type_id) && ga.gatefold_transparent)
	{
		g.gatefold_transparent = ga.gatefold_transparent
		//delete g.images.gatefold1
	}

	return g
}

export async function queryStats() : Promise<Array<StatsProps>>
{
	const env = process.env.NODE_ENV
	let cache
	if(redis)
	{
		cache = await redis.get("all_games_stats");
	}
	if(env == "production" && cache)
	{
		return JSON.parse(cache)
	}

	const games = await db.Software.findAll({
		attributes:['box_type_id','steam_link','gog_link','other_link']
	})
	
	let boxStats:Array<StatsProps> = [
		{ name:'Total Boxes',count:games.length },
		{ name:'(Probably) Abandonware',count:0, id:-1 }
	]
	forEach(games,function(g,i)
	{
		let bt = BigBoxTypes[g.box_type_id].replaceAll('_', ' ');
		if(findIndex(boxStats,{name:bt}) == -1)
		{
			boxStats.push({name:bt,count:1,id:g.box_type_id})
		}
		else
		{
			boxStats[findIndex(boxStats,{name:bt})].count += 1
		}

		if(!g.steam_link && !g.gog_link && !g.other_link)
		{
			boxStats[findIndex(boxStats,{name:'(Probably) Abandonware'})].count++
		}
	})

	if(redis)
	{
		await redis.set("all_games_stats", JSON.stringify(boxStats), "EX", 3600);
	}

	return boxStats
}

export async function getGameMetadata(slug:string)
{
	const env = process.env.NODE_ENV
	let cache
	if(redis)
	{
		cache = await redis.get("meta_"+slug);
	}
	if(env == "production" && cache)
	{
		return JSON.parse(cache)
	}

	let g = await queryGame({slug:slug},['slug', 'box_type_id', 'title'])	
	if(redis)
	{
		await redis.set("meta_"+slug, JSON.stringify(g), "EX", 3600);
	}

	return g
}

export async function getDeveloperMetadata(slug:string)
{
	const env = process.env.NODE_ENV
	let cache
	if(redis)
	{
		cache = await redis.get("metadev_"+slug);
	}
	if(env == "production" && cache)
	{
		return JSON.parse(cache)
	}

	let d = await db.Developer.findOne({
			where:{slug:slug}
	})
	if(redis)
	{
		await redis.set("metadev_"+slug, JSON.stringify(d), "EX", 3600);
	}

	return d
}

export async function getPublisherMetadata(slug:string)
{
	const env = process.env.NODE_ENV
	let cache
	if(redis)
	{
		cache = await redis.get("metadev_"+slug);
	}
	if(env == "production" && cache)
	{
		return JSON.parse(cache)
	}

	let d = await db.Publisher.findOne({
			where:{slug:slug}
	})
	if(redis)
	{
		await redis.set("metadev_"+slug, JSON.stringify(d), "EX", 3600);
	}

	return d
}

export async function getDevelopers()
{
	const env = process.env.NODE_ENV
	let cache
	if(redis)
	{
		cache = await redis.get("developer_list");
	}
	if(env == "production" && cache)
	{
		return JSON.parse(cache)
	}

	let d = await db.Developer.findAll({
		attributes: {
			include: [
				[fn('COUNT', col('software.id')), 'softwareCount'],
			],
			
		},
		include: [
			{
			model: db.Software,
				as: 'software',
				attributes: [],     // important: don't select software columns
				through: { attributes: [] }, // hide join table
			},
		],
		order:['name'],
		group: ['Developer.id'],
	})
	if(redis)
	{
		await redis.set("developer_list", JSON.stringify(d), "EX", 3600);
	}

	return d
}

export async function getPublishers()
{
	const env = process.env.NODE_ENV
	let cache
	if(redis)
	{
		cache = await redis.get("publisher_list");
	}
	if(env == "production" && cache)
	{
		return JSON.parse(cache)
	}

	let d = await db.Publisher.findAll({
		attributes: {
			include: [
				[fn('COUNT', col('software.id')), 'softwareCount'],
			],
		},
		include: [
			{
			model: db.Software,
				as: 'software',
				attributes: [],     // important: don't select software columns
				through: { attributes: [] }, // hide join table
			},
		],
		order:['name'],
		group: ['Publisher.id'],
	})
	if(redis)
	{
		await redis.set("publisher_list", JSON.stringify(d), "EX", 3600);
	}

	return d
}