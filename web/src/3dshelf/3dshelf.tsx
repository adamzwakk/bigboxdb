import Search from "@/partials/Search";
import MainShelves from "./MainShelves";
import ShelvesProvider from "./ShelvesProvider";
import Footer from "@/partials/Footer";

export default function ThreeDeeShelf(){
    return (
        <ShelvesProvider>
            <title>BigBoxDB | 3D Shelves</title>
            <div id="topLogo" className="threedee">
              <a href="/" className="logoContain mb-2 block">
                  <div className="mainLogo">
                      <img src="/img/logo_filled.png" alt="Logo"/>
                  </div>
                  <div className="mainLogoText">
                      <h1 className='font-bold'>BigBoxDB</h1>           
                      <h2 className='font-bold'>an elegant wrapping<br/>from a more civilized age</h2>
                  </div>
              </a>
              <Search onShelf={true} />
              {/* <ShelfOptions /> */}
          </div>
          {/* <GameInfo />
          <Music /> */}
          <div id="app">
            <MainShelves />
          </div>
          <div id="footer" className='fixed w-[100%] flex text-xs font-bold bottom-0 right-0 text-white z-5 justify-end'>
            <Footer />
          </div>
        </ShelvesProvider>
    )
}