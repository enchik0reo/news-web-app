import React from 'react';
import '../css/login.css';
import { Nav } from 'react-bootstrap';

const LoginFormSuccess = (props) => {
  return (
    <div>
        <div className="app-wrapper">
          <h2 className="answer-title">{props.answer}</h2>

          <div className="tologin">
            <Nav.Link href="/suggest" > 
              <button className="submit">Go</button>
            </Nav.Link>
          </div>
        </div>
    </div>
  )
}

export default LoginFormSuccess