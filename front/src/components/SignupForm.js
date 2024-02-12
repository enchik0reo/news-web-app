import React, { useEffect, useState } from 'react';
import validation from './Validation';
import axios from 'axios';
import { toast } from 'react-toastify';
import { Link } from 'react-router-dom';

const baseurl = "/signup"

const SignupForm = ({ submitForm }) => {

    const [values, setValues] = useState({
        fullname: "",
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
        setErrors(validation(values));
        setDataIsCorrect(true)
    }

    useEffect(() => {
        if (Object.keys(errors).length === 0 && dataIsCorrect) {

            const jsonData = {
                name: values.fullname,
                email: values.email,
                password: values.password
            };

            axios.post(baseurl, jsonData, {}).then((r) => {
                submitForm(r)
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
            <div className="app-wrapper">
                <div>
                    <h2 className="title">Create Account</h2>
                </div>
                <form className="form-wrapper">
                    <div className="name">
                        <label className="label">Name</label>
                        <input
                            className="input"
                            type="text"
                            name="fullname"
                            value={values.fullname}
                            onChange={handleChange}
                        />
                        {errors.fullname && <p className="error">{errors.fullname}</p>}
                    </div>

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
                            Sign Up
                        </button>
                    </div>
                </form>
            </div>
            <div className="app-mini-login">
                <p>Already have an account?</p>
                <Link className="back-link" to="/login">Login</Link>
            </div>
        </div>
    )
}

export default SignupForm