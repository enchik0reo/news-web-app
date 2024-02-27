const Validation = (value) => {

    let errors = {}

    if (!value) {
        errors.fullname = "Name is required."
    } else if (value.length > 30) {
        errors.fullname = "Name must be no more than 30 characters long."
    } else if (value.includes(' ')) {
        errors.fullname = "Name must not contain spaces."
    } else if (/\S+@\S+\.\S+/.test(value)) {
        errors.fullname = "Name can't be an E-mail."
    }


    return errors
}

export default Validation