import { BrowserRouter, Route, Routes } from 'react-router'
import { Outlet } from 'react-router';
import Home from './main/Home';
import ThreeDeeShelf from './3dshelf/3dshelf';
import Footer from './partials/Footer';
import Variant from './main/Variant';
import Faq from './main/Faq';
import Game from './main/Game';

function MainLayout() {
  return (
    <>
          <div id="bg" className="w-screen h-screen fixed z-0">
              <div className="bg-gradient w-screen h-screen absolute z-1"></div>
              <div className="bg-3d w-screen h-screen">
              <div className="bg-tiles w-6000 h-6000 bg-repeat absolute blur-xs opacity-80"></div>
              </div>
          </div>
          <div id="main-content" className="relative min-h-screen overflow-y-auto">
              <Outlet />
          </div>
        <div className="h-30">&nbsp;</div>
        <div id="footer" className='fixed w-[100%] flex text-xs font-bold bottom-0 text-white z-5 justify-center'>
            <Footer />
        </div>
    </>
  );
}

function App() {

  return (
    <BrowserRouter>
        <Routes>
            <Route path="/" element={<MainLayout />}>
                <Route index element={<Home />} />
                <Route path="/faq" element={<Faq />} />
                <Route path="/game/:gameSlug" element={<Game />} />
                <Route path="/game/:gameSlug/:variantId" element={<Variant />} />
            </Route>
            <Route path="/shelves">
                <Route index element={<ThreeDeeShelf />} />
                <Route path="/shelves/game/:gameSlug/:variantId" element={<ThreeDeeShelf />} />
            </Route>
        </Routes>
    </BrowserRouter>
  )
}

export default App
