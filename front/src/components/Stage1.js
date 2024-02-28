import React, { useEffect, useState } from 'react';
import validation from './Valid1';
import { Link } from 'react-router-dom';
import axios from 'axios';
import { toast } from 'react-toastify';
import { IoCheckmark } from 'react-icons/io5';

const baseurl = "/check/email"

const Stage1 = ({ values, onValues, errors, onErrors, dataIsCorrect, onDataIsCorrect, onSt2 }) => {

    const [doRequest, setDoRequest] = useState(false)
    const [clicked, setClicked] = useState(false)

    const handleChange = (event) => {
        onValues({
            ...values,
            [event.target.name]: event.target.value,
        })
    }

    const handleFormSubmit = (event) => {
        event.preventDefault()
        setClicked(true)
        onErrors(validation(values, setDoRequest))
        if (doRequest === true) {
            onDataIsCorrect(true)
        } else {
            onDataIsCorrect(false)
        }
    }

    useEffect(() => {
        if (Object.keys(errors).length === 0 && dataIsCorrect) {
            if (doRequest && values.email !== "" && !values.email.includes(' ')) {
                const jsonData = {
                    email: values.email,
                }

                axios.post(baseurl, jsonData, {}).then((res) => {
                    if (res.data.body.exists) {
                        toast.warn('User already exists.')
                        onDataIsCorrect(false)
                    } else {
                        onDataIsCorrect(true)
                        onErrors(errors)
                        onDataIsCorrect(dataIsCorrect)
                        onSt2(true)
                    }
                })
                    .catch((error) => {
                        console.error('Internal server error:', error)
                        toast.error('Internal server error. Please try later.')
                        onDataIsCorrect(false)
                    })
            }
        }
    }, [doRequest, errors, dataIsCorrect, values, onErrors, onDataIsCorrect, onSt2])

    return (
        <div>
            <div className="app-wrapper">
                <div>
                    <h2 className="title">Create Account</h2>
                </div>
                <form className="form-wrapper">
                    <div className="email">
                        <label className="label">E-mail</label>
                        {!errors.email && clicked ? <IoCheckmark className="checked" /> : <></>}
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
                        {!errors.password && clicked ? <IoCheckmark className="checked" /> : <></>}
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
                        {!errors.repassword && clicked ? <IoCheckmark className="checked" /> : <></>}
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

                    {doRequest
                        ? <div>
                            <button className="next-submit" onClick={handleFormSubmit}>
                                Next
                            </button>
                        </div>
                        : <div>
                            <button className="fix-submit" onClick={handleFormSubmit}>
                                Check fields
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