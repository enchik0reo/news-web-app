import React from 'react';
import '../css/pages.css';
import Forml from '../components/Forml';

export default class Login extends React.Component {

    render() {
        return (
            <>
                <Forml onLoginForm={this.props.onLoginForm} />
            </>
        )
    }
}