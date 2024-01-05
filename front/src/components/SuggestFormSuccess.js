import React from 'react';
import '../css/login.css';
import { Nav } from 'react-bootstrap';

const SuggestFormSuccess = (props) => {
  return (
    <div>
        <div className="app-wrapper">
          <h2 className="answer-sub">{props.answer}</h2>
          <h5 className="answer-sub">We will check it and if everything is ok, we'll publish it soon.</h5>
          <div className="tologin">
            <Nav.Link href="/" > 
              <button className="submit">Home</button>
            </Nav.Link>
          </div>
        </div>
    </div>
  )
}

export default SuggestFormSuccess