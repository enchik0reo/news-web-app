const Validation = (values) => {

    let errors = {}

    if (!values.email) {
        errors.email = "E-mail is required."
    } else if (!/\S+@\S+\.\S+/.test(values.email)) {
        errors.email = "E-mail is invalid."
    }  else if (values.email.includes(' ')) {
        errors.email = "E-mail must not contain spaces."
    }
    
    if (!values.password) {
        errors.password = "Password is required."
    } else if (values.password.length < 5) {
        errors.password = "Password must be at least 5 characters long."
    }  else if (values.password.includes(' ')) {
        errors.password = "Password must not contain spaces."
    }

    if (!values.repassword) {
        errors.repassword = "Required field."
    } else if (values.password !== values.repassword) {
        errors.repassword = "Password mismatch."
    }

    return errors
}

export default Validation