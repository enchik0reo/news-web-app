import React from 'react';
import { Nav } from 'react-bootstrap';
import '../css/suggest.css';

const NoLoggedForm = () => {

    return (
        <div className="suggest-app">
            <h2>Please login to offer news!</h2>
            <div className="tologin">
                <Nav.Link href="/login" >
                    <button className="submit">Login</button>
                </Nav.Link>
            </div>
        </div>
    )
}

export default NoLoggedForm