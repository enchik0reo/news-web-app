import React, { useState } from 'react';
import SignupForm from './SignupForm';
import SignupFormSuccess from './SignupFormSuccess';
import { toast } from 'react-toastify';

const Form = () => {

  const [formIsSubmitted, setFormIsSubmitted] = useState(false)
  const [answer, setAnswer] = useState('')

  const submitForm = (props) => {
    if (props.data.status === 201) {
      setAnswer('Account Created!')
      setFormIsSubmitted(true)
    } else if (props.data.status === 204) {
      setAnswer('User already exists!')
      setFormIsSubmitted(true)
    } else if (props.data.status === 400) {
      toast.warn("Name is empty.")
      setFormIsSubmitted(false)
    } else {
      toast.error("Failed to signup. Internal server error.")
      console.error(props.data.status)
    }
  }

  return (
    <div>
      {!formIsSubmitted ? <SignupForm submitForm={submitForm} /> : <SignupFormSuccess answer={answer} />}
    </div>
  )
}

export default Form