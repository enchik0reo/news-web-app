import React, { useState } from 'react';
import LoginForm from './LoginForm';
import LoginFormSuccess from './LoginFormSuccess';
import { toast } from 'react-toastify';

const Forml = ({ onLoginForm }) => {

  const [formIsSubmitted, setFormIsSubmitted] = useState(false)
  const [answer, setAnswer] = useState('')

  const submitForm = (props) => {
    if (props.data.status === 202) {
      setAnswer('You have successfully logged in!')
      localStorage.setItem('access_token', 'Bearer ' + props.data.body.access_token)
      setFormIsSubmitted(true)
      onLoginForm(true)
    } else {
      toast.error("Failed to login. Internal server error.")
      console.error(props.data.status)
    }
  }

  return (
    <div>
      {!formIsSubmitted ? <LoginForm submitForm={submitForm} /> : <LoginFormSuccess answer={answer} />}
    </div>
  )
}

export default Forml