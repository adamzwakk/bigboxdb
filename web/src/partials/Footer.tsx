export default function Footer()
{
    return(
        <ul className="bg-black/50 p-2">
            <ul className="text-center">
                <li className="inline mr-2"><a href="/faq" className="underline">FAQ</a></li>
                <li className="inline socialIcons">
                    <a rel="external" target='_blank' href="https://discord.gg/CWXbgxCeW5">
                        <span>
                            <object type="image/svg+xml" data='/img/icons/discord.svg' className='inline w-5 mr-2 pointer-events-none'></object>
                        </span>
                    </a>
                    <a rel="external" target='_blank' href="https://bsky.app/profile/bigboxdb.com">
                        <span>
                            <object type="image/svg+xml" data='/img/icons/bluesky.svg' className='inline w-5 pointer-events-none'></object>
                        </span>
                    </a>
                </li>
            </ul>
            <li>Created by <a href="https://www.adamzwakk.com" target="_blank" className="underline">Uncle Hans</a></li>
        </ul>
    )
}