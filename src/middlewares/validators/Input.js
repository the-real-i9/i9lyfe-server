import {
  isDate,
  isEmail,
  isEmpty,
  isNumeric,
} from "validator"

export default class Input {
  constructor(field, value) {
    this.field = field
    this.value = value
    this.error = null
  }

  notEmpty() {
    if (this.error) return this
    if (
      this.value === undefined ||
      this.value === null ||
      !isEmpty(this.value)
    ) {
      this.error = { field: this.field, msg: `${this.field} is required` }
    }

    return this
  }

  isEmail() {
    if (this.error) return this
    if (!isEmail(this.value)) {
      this.error = { field: this.field, msg: "invalid email" }
    }

    return this
  }

  isNumeric() {
    if (this.error) return this
    if (!isNumeric(this.value, { no_symbols: true })) {
      this.error = {
        field: this.field,
        msg: `${this.field}'s value is not a number`,
      }
    }

    return this
  }

  min(minValue) {
    if (this.error) return this
    if (this.value < minValue) {
      this.error = {
        field: this.field,
        msg: `${this.field}'s value is too short`,
      }
    }

    return this
  }

  isValidUsername() {
    if (this.error) return this
    if (!/^[\w-]{3,}$/.test(this.value)) {
      this.error = {
        field: this.field,
        msg: `${this.field}'s value contains unwanted characters`,
      }
    }

    return this
  }

  isDate() {
    if (this.error) return this
    if (!isDate(this.value)) {
      this.error = {
        field: this.field,
        msg: `${this.field}'s value must be a date string`,
      }
    }

    return this
  }
}
