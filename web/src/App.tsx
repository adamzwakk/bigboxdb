import { BrowserRouter } from 'react-router'
import MainShelves from './3dshelf/MainShelves'
import ShelvesProvider from './3dshelf/ShelvesProvider'

function App() {

  return (
    <ShelvesProvider>
        <BrowserRouter>
            <MainShelves />
        </BrowserRouter>
    </ShelvesProvider>
  )
}

export default App
