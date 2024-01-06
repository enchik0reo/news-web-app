import React, { useState } from 'react';
import LoginForm from './LoginForm';
import LoginFormSuccess from './LoginFormSuccess';

const Forml = ({ onLoginForm }) => {
  
  const [formIsSubmitted, setFormIsSubmitted] = useState(false)
  const [answer, setAnswer] = useState('')

  const submitForm = (props) => {
    if (props.status === 202) {
      setAnswer('You have successfully logged in!')
      localStorage.setItem('access_token', 'Bearer ' + props.headers.access_token)
      setFormIsSubmitted(true)
      onLoginForm(true)
    } else {
      console.log(props.status)
    }
  }

  return (
    <div>
      { !formIsSubmitted ? <LoginForm submitForm={submitForm}/> : <LoginFormSuccess answer={answer}/>}
    </div>
  )
}

export default Forml