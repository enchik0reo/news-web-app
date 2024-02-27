import React, { useEffect } from 'react';
import validation from './Valid1';
import { Link } from 'react-router-dom';
import axios from 'axios';
import { toast } from 'react-toastify';

const baseurl = "/check/email"

const Stage1 = ({ values, onValues, errors, onErrors, dataIsCorrect, onDataIsCorrect, onSt2 }) => {

    const handleChange = (event) => {
        onValues({
            ...values,
            [event.target.name]: event.target.value,
        })
    }

    const handleFormSubmit = (event) => {
        event.preventDefault()
        onErrors(validation(values))

        if (Object.keys(errors).length === 0 && values.email !== "" && !values.email.includes(' ')) {
            const jsonData = {
                email: values.email,
            }
            axios.post(baseurl, jsonData, {}).then((res) => {
                if (res.data.body.exists) {
                    toast.warn('User already exists')
                    onDataIsCorrect(false)
                } else {
                    onDataIsCorrect(true)
                }
            })
                .catch((error) => {
                    console.error('Internal server error:', error)
                    toast.error('Can`t check e-mail. Internal server error.')
                    onDataIsCorrect(false)
                })
        }
    }

    useEffect(() => {
        if (Object.keys(errors).length === 0 && dataIsCorrect) {
            onErrors(errors)
            onDataIsCorrect(dataIsCorrect)
            onSt2(true)
        }
    }, [errors, dataIsCorrect, values, onErrors, onDataIsCorrect, onSt2])

    return (
        <div>
            <div className="app-wrapper">
                <div>
                    <h2 className="title">Create Account</h2>
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

                    <div className="repassword">
                        <label className="label">Repeat Password</label>
                        <input
                            id="pass"
                            className="input"
                            type="password"
                            name="repassword"
                            value={values.repassword}
                            onChange={handleChange}
                        />
                        {errors.repassword && <p className="error">{errors.repassword}</p>}
                    </div>

                    {errors.email || errors.password || errors.repassword
                        ? <div>
                            <button className="fix-submit" onClick={handleFormSubmit}>
                                Fix
                            </button>
                        </div>
                        : <div>
                            <button className="yes-submit" onClick={handleFormSubmit}>
                                Next
                            </button>
                        </div>}
                </form>
            </div>
            <div className="app-mini-login">
                <p>Already have an account?</p>
                <Link className="back-link" to="/login">Login</Link>
            </div>
        </div>
    )
}

export default Stage1