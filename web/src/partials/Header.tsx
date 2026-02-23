import Search from "./Search";

export default function Header()
{
    // This is used on all regular non-shelf pages except the home page
    return (
        <div id="main-header" className="sm:mb-8 mb-4 bg-black/35 px-10 py-2 z-99 relative sm:flex flex-none w-full text-white justify-stretch border-b-1 border-b-white-200">
            <a href="/" className="flex-none sm:text-[34px] text-[36px] text-center block font-bold">
                <img className='mainLogo inline w-[32px] mr-2 mb-2' src="/img/logo_filled.png" alt="Logo"/>&nbsp;<span>BigBoxDB</span>
                {/* <h2 className='sm:text-[10px] text-[14px] -mt-2 text-center block font-bold'>an elegant wrapping from a more civilized age</h2> */}
            </a>
            <div className="flex-3 sm:ml-10 mt-2">
                <div className="sm:w-xs">
                    <Search onShelf={false} />
                </div>
            </div>
            <div>
                <ul className='sm:mt-1 mt-3 text-center'>
                    <li className='inline-block sm:bg-black/35 bg-none sm:p-3 p-1 border-1'>
                        <a href="/shelves"><span className='font-bold'>3D Shelves</span></a>
                    </li>
                    <li className='inline-block ml-2 sm:bg-black/35 bg-none sm:p-3 p-1 border-1'>
                        <a href="/faq" className='font-bold'>FAQ</a>
                    </li>
                </ul>
            </div>
        </div>
    )
}