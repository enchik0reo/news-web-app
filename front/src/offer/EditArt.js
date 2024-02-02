import React, { useEffect, useState } from 'react';
import 'react-toastify/dist/ReactToastify.css';
import '../css/suggest.css'
import valid from '../components/Valid';
import { Nav } from 'react-bootstrap';

const EditArt = ({ id, onEdit }) => {

    const [values, setValues] = useState({
        link: "",
        article_id: id,
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
            onEdit(values)
            setDataIsCorrect(false)
        }
    }, [errors, dataIsCorrect, onEdit, values])

    return (
        <div className="offer-edit-app">
            <form className="form-wrapper">
                <div className="email">
                    <label className="label">New link to article</label>
                    <input
                        className="input"
                        type="link"
                        name="link"
                        value={values.link}
                        onChange={handleChange}
                    />
                    {errors.link && <p className="error">{errors.link}</p>}
                </div>

                <div className="tologin">
                    <Nav.Link>
                        <button className="change" onClick={handleFormSubmit}>Change</button>
                    </Nav.Link>
                </div>
            </form>
        </div>
    )
}

export default EditArt