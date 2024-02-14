import React from 'react';
import './css/main.css';
import 'bootstrap/dist/css/bootstrap.min.css';
import Header from './components/Header';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

const App = () => {

  return (
    <div>
      <Header />
      <ToastContainer />
    </div>
  )
}

export default App
