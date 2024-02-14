const Valid = (values) => {

  let errors = {}

  if (!values.link) {
    errors.link = "Link is required."
  }

  return errors
}

export default Valid