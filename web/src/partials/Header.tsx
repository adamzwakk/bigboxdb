import Search from "./Search";

export default function Header()
{
    // This is used on all regular non-shelf pages except the home page
    return (
        <div id="main-header" className="sm:flex flex-none ml-auto mr-auto w-full sm:w-2xl p-5 z-4 text-white">
            <a href="/" className="mr-10 flex-none">
                <h1 className='sm:text-[34px] text-[36px] text-center block font-bold'><img className='mainLogo inline w-[32px] mr-2 mb-2' src="/img/logo_filled.png" alt="Logo"/> BigBoxDB</h1>
                {/* <h2 className='sm:text-[10px] text-[14px] -mt-2 text-center block font-bold'>an elegant wrapping from a more civilized age</h2> */}
            </a>
            <div className="flex-3">
                <Search onShelf={false} />
                <ul className='text-right belowSearch'>
                    <li className='inline-block bg-black/50 p-2'>
                        <a href="/shelves"><img src="/img/icons/shelves.png" className="inline w-6" alt="" />Or see the <span className='font-bold underline'>3D shelves!</span></a>
                    </li>
                </ul>
            </div>
        </div>
    )
}