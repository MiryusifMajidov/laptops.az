import { useEffect } from 'react'
import { Routes, Route } from 'react-router-dom'
import Layout from './Layout'
import Home from './pages/Home'
import Category from './pages/Category'
import Product from './pages/Product'
import { trackVisit } from './api'

export default function App() {
  useEffect(() => { trackVisit() }, []) // sayt açılanda ziyarəti logla (sessiya başına bir dəfə)
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Home />} />
        <Route path="/mehsullar" element={<Category />} />
        <Route path="/kateqoriya/:name" element={<Category />} />
        <Route path="/mehsul/:id" element={<Product />} />
        <Route path="*" element={<Home />} />
      </Route>
    </Routes>
  )
}
