import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import validationl from './Validationl';
import axios from 'axios';
import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

const baseurl = "/login"

const LoginForm = ({ submitForm }) => {

    const [values, setValues] = useState({
        email: "",
        password: "",
    })

    const [errors, setErrors] = useState({});
    const [dataIsCorrect, setDataIsCorrect] = useState(false)

    const handleChange = (event) => {
        setValues({
            ...values,
            [event.target.name]: event.target.value,
        })
    }

    const handleFormSubmit = (event) => {
        event.preventDefault();
        setErrors(validationl(values));
        setDataIsCorrect(true)
    }

    useEffect(() => {
        if (Object.keys(errors).length === 0 && dataIsCorrect) {

            const jsonData = {
                email: values.email,
                password: values.password
            };

            axios.post(baseurl, jsonData, {}).then((r) => {
                if (r.data.status === 204) {
                    toast.warn("Wrong e-mail or password.")
                    setDataIsCorrect(false)
                } else {
                    submitForm(r)
                }
            })
                .catch((error) => {
                    if (error) {
                        toast.error("Internal server error. Please, try later.")
                        console.error('Internal server error:', error)
                        setDataIsCorrect(false)
                    }
                })
        }
    }, [errors, dataIsCorrect, submitForm, values])

    return (
        <div>
            <div className="app-wrapperl">
                <div>
                    <h2 className="titlel">Login to Account</h2>
                </div>
                <form className="form-wrapper">
                    <div className="email">
                        <label className="label">E-mail</label>
                        <input
                            className="input"
                            type="email"
                            name="email"
                            value={values.email}
                            onChange={handleChange}
                        />
                        {errors.email && <p className="error">{errors.email}</p>}
                    </div>

                    <div className="password">
                        <label className="label">Password</label>
                        <input
                            className="input"
                            type="password"
                            name="password"
                            value={values.password}
                            onChange={handleChange}
                        />
                        {errors.password && <p className="error">{errors.password}</p>}
                    </div>

                    <div>
                        <button className="submit" onClick={handleFormSubmit}>
                            Sign In
                        </button>
                    </div>
                </form>
            </div>
            <div className="app-mini">
                <p>Don't have an account?</p>
                <Link to="/signup">Create</Link>
            </div>
        </div>
    )
}

export default LoginForm