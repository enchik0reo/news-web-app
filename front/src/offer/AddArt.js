import React, { useEffect, useState } from 'react';
import 'react-toastify/dist/ReactToastify.css';
import '../css/suggest.css'
import valid from '../components/Valid';
import { Nav } from 'react-bootstrap';

const AddArt = ({ onAdd }) => {

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
            onAdd(values)
            setDataIsCorrect(false)
        }
    }, [errors, dataIsCorrect, onAdd, values])

    return (
        <div className="offer-app">
            <div>
                <h3 className="offer-title">Suggest Article</h3>
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
                    <textarea className="offer-input-text"
                        type="content"
                        name="content"
                        value={values.content}
                        onChange={handleChange} />
                    {errors.content && <p className="error">{errors.content}</p>}
                </div>

                <div className="tologin">
                    <Nav.Link>
                        <button className="offer-submit" onClick={handleFormSubmit}>Send</button>
                    </Nav.Link>
                </div>
            </form>
        </div>
    )
}

export default AddArt