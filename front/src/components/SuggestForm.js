import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import '../css/suggest.css'
import valid from './Valid';
import { Nav } from 'react-bootstrap';

const baseurl = "/suggest"

const SuggestForm = ({ submitForm }) => {

    const [values, setValues] = useState({
        link: "",
        content: "",
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
        setErrors(valid(values));
        setDataIsCorrect(true)
    }

    useEffect(() => {
        if (Object.keys(errors).length === 0 && dataIsCorrect) {

            const jsonData = {
                link: values.link,
                content: values.content
            };

            const config = {
                headers: {
                    "Authorization": localStorage.getItem('access_token'),
                }
            }

            axios.post(baseurl, jsonData, config).then((r) => {
                if (r.status === 206) {
                    toast("We already know about this article and will publish it soon. Thank you!")
                    setDataIsCorrect(false)
                } else if (r.status === 204) {
                    toast("This article is not suitable, sorry. Please, try another one.")
                    setDataIsCorrect(false)
                } else {
                    submitForm(r)
                }
            })
                .catch((error) => {
                    if (error) {
                        if (error.response.statusText) {
                            if (error.response.statusText === "Unauthorized") {
                                toast("Sorry, your session is expired. Please, relogin.")
                            }
                        }
                        console.error('Ошибка при выполнении запроса:', error)
                        setDataIsCorrect(false)
                    }
                })

        }
    }, [errors, dataIsCorrect, submitForm, values])

    return (
        <div className="suggest-appp">
            <div>
                <h2 className="titlel">Suggest Article</h2>
            </div>
            <form className="form-wrapper">
                <div className="email">
                    <label className="label">Link to article</label>
                    <input
                        className="input"
                        type="link"
                        name="link"
                        value={values.link}
                        onChange={handleChange}
                    />
                    {errors.link && <p className="error">{errors.link}</p>}
                </div>

                <div className="password">
                    <label className="label">Short description (optional)</label>
                    <textarea className="input-text"
                        type="content"
                        name="content"
                        value={values.content}
                        onChange={handleChange} />
                    {errors.content && <p className="error">{errors.content}</p>}
                </div>

                <div className="tologin">
                    <Nav.Link>
                        <button className="submit" onClick={handleFormSubmit}>Send</button>
                    </Nav.Link>
                </div>
            </form>
        </div>
    )
}

export default SuggestForm