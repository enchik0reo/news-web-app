import React, { useState, useEffect } from 'react';
import '../css/pages.css';
import NoLoggedForm from '../components/NoLoggedForm';
import LoggedForm from '../offer/LoggedForm';

const Offer = () => {

    const [formIsLogged, setFormIsLogged] = useState(false)

    useEffect(() => {
        if (localStorage.getItem('access_token') !== null) {
            setFormIsLogged(true)
        } else {
            setFormIsLogged(false)
        }
    }, []);
    
    return (
        <>
            {!formIsLogged ? <NoLoggedForm /> : <LoggedForm />}
        </>
    )
}

export default Offer