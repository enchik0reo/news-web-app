import React from 'react';
import '../css/signup.css';
import { Nav } from 'react-bootstrap';

const SignupFormSuccess = (props) => {
  return (
    <div>
      <div className="app-wrapper">
        <h1 className="answer-title">{props.answer}</h1>

        <div className="tologin">
          <Nav.Link href="/login" >
            <button className="submit">Login</button>
          </Nav.Link>
        </div>
      </div>
    </div>
  )
}

export default SignupFormSuccess