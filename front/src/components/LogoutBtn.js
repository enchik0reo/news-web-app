import React from 'react';
import { Nav } from 'react-bootstrap';


const LogoutBtn = ( onLoginBtn ) => {

    const logoutDo = () => {
        localStorage.removeItem('access_token')
        localStorage.removeItem('user_name')
        onLoginBtn(true)
    }

    return (
        <Nav.Item>
            <Nav.Link className="btn btn-outline-info" href="/" onClick={logoutDo} > Logout ({localStorage.getItem('user_name')}) </Nav.Link>
        </Nav.Item>
    )
}

export default LogoutBtn