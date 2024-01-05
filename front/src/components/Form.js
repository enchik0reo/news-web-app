import React, { useState } from 'react';
import SignupForm from './SignupForm';
import SignupFormSuccess from './SignupFormSuccess';

const Form = () => {
  
  const [formIsSubmitted, setFormIsSubmitted] = useState(false)
  const [answer, setAnswer] = useState('')

  const submitForm = (props) => {
    if (props.status === 201) {
      setAnswer('Account Created!')
      setFormIsSubmitted(true)
    } else if (props.status === 204) {
      setAnswer('User already exists!')
      setFormIsSubmitted(true)
    } else {
      console.log(props.status)
    }
  }

  return (
    <div>
      { !formIsSubmitted ? <SignupForm submitForm={submitForm}/> : <SignupFormSuccess answer={answer}/>}
    </div>
  )
}

export default Form