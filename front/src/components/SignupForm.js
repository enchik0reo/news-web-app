import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { toast } from 'react-toastify';
import Stage1 from './Stage1';
import Stage2 from './Stage2';

const baseurl = "/signup"

const SignupForm = ({ submitForm }) => {

    const [St2, setSt2] = useState(false)

    const [values1, setValues1] = useState({
        email: "",
        password: "",
        repassword: "",
    })
    const [values2, setValues2] = useState({
        fullname: "",
    })

    const [errors1, setErrors1] = useState({})
    const [errors2, setErrors2] = useState({})

    const [dataIsCorrect1, setDataIsCorrect1] = useState(false)
    const [dataIsCorrect2, setDataIsCorrect2] = useState(false)

    useEffect(() => {
        if (Object.keys(errors1).length === 0 && dataIsCorrect1 && Object.keys(errors2).length === 0 && dataIsCorrect2) {

            const jsonData = {
                name: values2.fullname,
                email: values1.email,
                password: values1.password
            }

            axios.post(baseurl, jsonData, {}).then((r) => {
                submitForm(r)
            })
                .catch((error) => {
                    if (error) {
                        toast.error("Internal server error. Please, try later.")
                        console.error('Internal server error:', error)
                        setDataIsCorrect1(false)
                        setDataIsCorrect2(false)
                        setValues1({
                            email: "",
                            password: "",
                            repassword: "",
                        })
                        setValues2({
                            fullname: "",
                        })
                        setSt2(false)
                    }
                })
        }
    }, [submitForm, errors1, errors2, dataIsCorrect1, dataIsCorrect2, values1, values2, St2])

    return (
        <>
            {!St2
                ? <Stage1 values={values1} onValues={setValues1} errors={errors1} onErrors={setErrors1} dataIsCorrect={dataIsCorrect1} onDataIsCorrect={setDataIsCorrect1} onSt2={setSt2} />
                : <Stage2 values={values2} onValues={setValues2} errors={errors2} onErrors={setErrors2} dataIsCorrect={dataIsCorrect2} onDataIsCorrect={setDataIsCorrect2} />}
        </>
    )
}

export default SignupForm