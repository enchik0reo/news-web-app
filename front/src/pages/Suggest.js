import React, { useState, useEffect } from 'react';
import '../css/pages.css';
import NoLoggedForm from '../components/NoLoggedForm';
import SuggestForm from '../components/SuggestForm';
import SuggestFormSuccess from '../components/SuggestFormSuccess';

const Suggest = () => {

    const [formIsLogged, setFormIsLogged] = useState(false)

    useEffect(() => {
        if (localStorage.getItem('access_token') !== null) {
            setFormIsLogged(true)
        } else {
            setFormIsLogged(false)
        }
    }, []);

    const [formIsSubmitted, setFormIsSubmitted] = useState(false)
    const [answer, setAnswer] = useState('')

    const submitForm = (props) => {
        if (props.status === 201) {
            setAnswer('You have successfully submitted an article!')
            if (props.headers.access_token) {
                localStorage.setItem('access_token', 'Bearer ' + props.headers.access_token)
            }
            setFormIsSubmitted(true)
        } else {
            console.log(props.status)
        }
    }

    return (
        <>
            {!formIsLogged ? <NoLoggedForm /> : !formIsSubmitted ? <SuggestForm submitForm={submitForm} /> : <SuggestFormSuccess answer={answer} />}
        </>
    )
}

export default Suggest