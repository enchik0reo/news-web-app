import React, { useState, useEffect } from 'react';
import LoginBtn from './LoginBtn';
import LogoutBtn from './LogoutBtn';

const SignBtn = () => {

    const [loginB, setLoginB] = useState(true)

    useEffect(() => {
        if (localStorage.getItem('access_token') !== null) {
            setLoginB(false)
        } else {
            setLoginB(true)
        }
    }, []);

    return (
        <div>
            {loginB ? <LoginBtn /> : <LogoutBtn />}
        </div>
    )
}

export default SignBtn