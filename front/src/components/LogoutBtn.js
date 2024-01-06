import React from 'react';
import { Nav } from 'react-bootstrap';


const LogoutBtn = () => {

    const logoutDo = () => {
        localStorage.removeItem('access_token')
    }

    return (
        <Nav.Item>
            <Nav.Link className="btn btn-outline-info" href="/" onClick={logoutDo} > Logout </Nav.Link>
        </Nav.Item>
    )
}

export default LogoutBtn