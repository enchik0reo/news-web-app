const Validation = (values, setDoRequest) => {

    let errors = {}

    if (!values.email) {
        errors.email = "E-mail is required."
        setDoRequest(false)
    } else if (!/\S+@\S+\.\S+/.test(values.email)) {
        errors.email = "E-mail is invalid."
        setDoRequest(false)
    } else if (values.email.includes(' ')) {
        errors.email = "E-mail must not contain spaces."
        setDoRequest(false)
    }

    if (!values.password) {
        errors.password = "Password is required."
        setDoRequest(false)
    } else if (values.password.length < 5) {
        errors.password = "Password must be at least 5 characters long."
        setDoRequest(false)
    } else if (values.password.includes(' ')) {
        errors.password = "Password must not contain spaces."
        setDoRequest(false)
    }

    if (!values.repassword) {
        errors.repassword = "Required field."
        setDoRequest(false)
    } else if (values.password !== values.repassword) {
        errors.repassword = "Password mismatch."
        setDoRequest(false)
    }

    if (errors.email === undefined && errors.password === undefined && errors.repassword === undefined) {
        setDoRequest(true)
    }

    return errors
}

export default Validation