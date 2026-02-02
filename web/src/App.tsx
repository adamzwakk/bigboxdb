import { BrowserRouter, Route, Routes } from 'react-router'
import MainShelves from './3dshelf/MainShelves'
import ShelvesProvider from './3dshelf/ShelvesProvider'

function App() {

  return (
    <ShelvesProvider>
        <BrowserRouter>
            <Routes>
                <Route path="/shelves" element={<MainShelves />} />
            </Routes>
        </BrowserRouter>
    </ShelvesProvider>
  )
}

export default App
