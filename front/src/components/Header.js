import React from 'react';
import { Navbar, Nav, Container } from 'react-bootstrap';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import logo from '../img/logo.png';
import Home from '../pages/Home';
import Suggest from '../pages/Suggest';
import Login from '../pages/Login';
import Signup from '../pages/Signup';
import SignBtn from './Sign';

const Header = () => {
    return (
        <div>
            <Navbar fixed="top" collapseOnSelect expand="sm" bg="dark" variant="dark" >
                <Container>
                    <Navbar.Brand href="/">
                        <img
                            src={logo}
                            height="30"
                            width="80"
                            className="d-inline-block align-top"
                            alt="logo"
                        /> Newsline
                    </Navbar.Brand>
                    <Navbar.Toggle aria-controls="responsive-navbar-nav" />
                    <Navbar.Collapse id="responsive-navbar-nav" >
                        <Nav className="me-auto">
                            <Nav.Link className="ms-4" href="/suggest" > Suggest  News </Nav.Link>
                        </Nav>
                        <Nav className="d-flex">
                            <SignBtn />
                        </Nav>
                    </Navbar.Collapse>
                </Container>
            </Navbar>

            <Router>
                <Routes>
                    <Route path="/" element={<Home />} />
                    <Route path="/suggest" element={<Suggest />} />
                    <Route path="/login" element={<Login />} />
                    <Route path="/signup" element={<Signup />} />
                </Routes>
            </Router>
        </div>
    )
}

export default Header