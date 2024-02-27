import React, { useEffect, useState } from 'react';
import validation from './Valid2';
import { Link } from 'react-router-dom';
import axios from 'axios';
import { toast } from 'react-toastify';

const baseurl = "/check/user_name"

const Stage2 = ({ values, onValues, errors, onErrors, dataIsCorrect, onDataIsCorrect }) => {

    const [activeBtn, setActiveBtn] = useState(false)
    const [clicked, setClicked] = useState(false)
    const [changed, setChanged] = useState(false)
    const [postErr, setPostErr] = useState("")

    const handleChange = (event) => {
        onValues({
            ...values,
            [event.target.name]: event.target.value,
        })

        onErrors(validation(event.target.value))

        if ((!/\S+@\S+\.\S+/.test(event.target.value)) && event.target.value !== "" && !event.target.value.includes(' ')) {
            const jsonData = {
                user_name: event.target.value,
            }

            axios.post(baseurl, jsonData, {}).then((res) => {
                if (res.data.body.exists) {
                    setPostErr("Name is already taken.")
                } else {
                    setPostErr("")
                }
            })
                .catch((error) => {
                    console.error('Internal server error:', error)
                    toast.error('Can`t check user name. Internal server error.')
                })
        } else {
            setPostErr("")
        }

        onDataIsCorrect(false)
        setChanged(true)
    }

    useEffect(() => {
        if (Object.keys(errors).length === 0 && postErr === "") {
            if (clicked || changed) {
                setActiveBtn(true)
            }
        } else {
            setActiveBtn(false)
        }
    }, [activeBtn, clicked, changed, errors, postErr])

    const handleFormSubmit = (event) => {
        event.preventDefault()
        if (activeBtn === true) {
            onErrors(validation(values.fullname))
            event.preventDefault()
            onDataIsCorrect(true)
            setClicked(true)
        }
    }

    useEffect(() => {
        if (Object.keys(errors).length === 0 && dataIsCorrect && postErr === "" && activeBtn === true) {
            onErrors(errors)
            onDataIsCorrect(dataIsCorrect)
        }
    }, [errors, dataIsCorrect, onErrors, onDataIsCorrect, postErr, activeBtn])

    return (
        <div>
            <div className="app-wrapper-2">
                <div>
                    <h2 className="title">Сhoose user name</h2>
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
                        {postErr !== "" ? <p className="error">{postErr}</p> : <></>}
                    </div>

                    {activeBtn
                        ? <div>
                            <button className="yes-submit" onClick={handleFormSubmit}>
                                Sign Up
                            </button>
                        </div>
                        : <div>
                            <button className="no-submit" onClick={handleFormSubmit}>
                                Sign Up
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

export default Stage2