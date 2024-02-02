import React from "react"
import UserArts from "./UserArts"
import AddArt from "./AddArt"
import axios from "axios"
import { toast } from 'react-toastify';
import "../css/offer.css"

const baseurl = "/user_news"

export default class LoggedForm extends React.Component {
  constructor(props) {
    super(props)

    const config = {
      headers: {
        "Authorization": localStorage.getItem('access_token'),
      }
    }

    axios.get(baseurl, config).then((res) => {
      if (res.headers.access_token) {
        localStorage.setItem('access_token', 'Bearer ' + res.headers.access_token)
      }

      if (res.status === 200) {
        this.setState({ articles: res.data })
      }
    })
      .catch((error) => {
        if (error) {
          toast.error("Internal server error. Failed to load suggested articles.")
          console.error('Ошибка при выполнении запроса:', error)
        }
      })

    this.state = {
      articles: []
    }

    this.deleteArticle = this.deleteArticle.bind(this)
    this.editArticle = this.editArticle.bind(this)
    this.addArticle = this.addArticle.bind(this)
  }
  render() {
    return (<div>
      <h3 className="special-offer-h3">Submit your articles!</h3>
      <main>
        <UserArts articles={this.state.articles} onEdit={this.editArticle} onDelete={this.deleteArticle} />
      </main>
      <aside>
        <AddArt onAdd={this.addArticle} />
      </aside>
    </div>)
  }

  deleteArticle(id) {
    
    const config = {
      headers: {
        "Authorization": localStorage.getItem('access_token'),
        "article_id": id,
      }
    }

    axios.delete(baseurl, config).then((res) => {
      if (res.headers.access_token) {
        localStorage.setItem('access_token', 'Bearer ' + res.headers.access_token)
      }
      if (res.status === 204) {
        toast.warn('You have successfully deleted an article! It won`t be published.')
        this.setState({ articles: [] })
      } else if (res.status === 200) {
        toast.warn('You have successfully deleted an article! It won`t be published.')
        this.setState({ articles: res.data })
      }
    })
      .catch((error) => {
        if (error) {
          toast.error("Internal server error.")
          console.error('Ошибка при выполнении запроса:', error)
        }
      })
  }

  editArticle(data) {

    const jsonEditData = {
      link: data.link,
      article_id: data.article_id
    };

    const config = {
      headers: {
        "Authorization": localStorage.getItem('access_token'),
      }
    }

    axios.put(baseurl, jsonEditData, config).then((res) => {
      if (res.headers.access_token) {
        localStorage.setItem('access_token', 'Bearer ' + res.headers.access_token)
      }
      if (res.status === 206) {
        toast.info("We already know about this article and will publish it soon. Thank you!")
      } else if (res.status === 204) {
        toast.warn("This article is not suitable, sorry. Please, try another one.")
      } else if (res.status === 205) {
        toast.info('You have successfully changed an article!')
        this.setState({ articles: [] })
      } else if (res.status === 202) {
        toast.info('You have successfully changed an article!')
        this.setState({ articles: res.data })
      }
    })
      .catch((error) => {
        if (error) {
          if (error.res && error.res.status === 401) {
            toast.error("Sorry, your session is expired. Please, relogin.")
          } else {
            toast.error("Internal server error. Failed to load suggested articles.")
            console.error('Ошибка при выполнении запроса:', error)
          }
        }
      })
  }

  addArticle(data) {
    const jsonAddData = {
      link: data.link,
      article_id: data.article_id
    };

    const config = {
      headers: {
        "Authorization": localStorage.getItem('access_token'),
      }
    }

    axios.post(baseurl, jsonAddData, config).then((res) => {
      if (res.headers.access_token) {
        localStorage.setItem('access_token', 'Bearer ' + res.headers.access_token)
      }
      if (res.status === 206) {
        toast.info("We already know about this article and will publish it soon. Thank you!")
      } else if (res.status === 204) {
        toast.warn("This article is not suitable, sorry. Please, try another one.")
      } else if (res.status === 205) {
        toast.success('You have successfully suggested an article!')
        this.setState({ articles: [] })
      } else if (res.status === 201) {
        toast.success('You have successfully suggested an article!')
        this.setState({ articles: res.data })
      }
    })
      .catch((error) => {
        if (error) {
          if (error.res && error.res.status === 401) {
            toast.error("Sorry, your session is expired. Please, relogin.")
          } else {
            toast.error("Internal server error. Failed to load suggested articles.")
            console.error('Ошибка при выполнении запроса:', error)
          }
        }
      })
  }
}