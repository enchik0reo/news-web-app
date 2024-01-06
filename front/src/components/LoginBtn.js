import React from 'react';
import { Nav } from 'react-bootstrap';

const LoginBtn = () => {
  return (
    <Nav.Item>
      <Nav.Link className="btn btn-outline-info" href="/login" > Login </Nav.Link>
    </Nav.Item>
  )
}

export default LoginBtn