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
      if (res.data.body.access_token) {
        localStorage.setItem('access_token', 'Bearer ' + res.data.body.access_token)
      }

      if (res.data.status === 200) {
        this.setState({ articles: res.data.body.articles })
      } else if (res.data.status === 204) {
        this.setState({ articles: [] })
      } else if (res.data.status === 401) {
        toast.error("Sorry, your session is wrong. Please, relogin.")
      } else if (res.data.status === 404) {
        toast.error('Please, login.')
      }
    })
      .catch((error) => {
        if (error) {
          toast.error("Internal server error. Failed to load suggested articles.")
          console.error('Internal server error:', error)
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
      <h3 className="special-offer-h3">Submit your articles</h3>
      <main>
        <UserArts articles={this.state.articles} onEdit={this.editArticle} onDelete={this.deleteArticle} />
      </main>
      <aside>
        <AddArt onAdd={this.addArticle} />
      </aside>
    </div>)
  }

  deleteArticle(id) {

    var result = window.confirm('Remove article?')
    if (!result) {
      return
    }

    const config = {
      headers: {
        "Authorization": localStorage.getItem('access_token'),
      },
      data: {
        "article_id": id,
      }
    }

    axios.delete(baseurl, config).then((res) => {
      if (res.data.body.access_token) {
        localStorage.setItem('access_token', 'Bearer ' + res.data.body.access_token)
      }
      if (res.data.status === 500) {
        toast.error("Internal server error. Please try later.")
      } else if (res.data.status === 401) {
        toast.error("Sorry, your session is wrong. Please, relogin.")
      } else if (res.data.status === 404) {
        toast.error('Please, login.')
      } else if (res.data.status === 204) {
        toast.info('You have successfully deleted an article!')
        this.setState({ articles: [] })
      } else if (res.data.status === 200) {
        toast.info('You have successfully deleted an article!')
        this.setState({ articles: res.data.body.articles })
      } else if (res.data.status === 208) {
        axios.get(baseurl, config).then((res) => {
          if (res.data.body.access_token) {
            localStorage.setItem('access_token', 'Bearer ' + res.data.body.access_token)
          }
          
          if (res.data.status === 200) {
            this.setState({ articles: res.data.body.articles })
          } else if (res.data.status === 204) {
            this.setState({ articles: [] })
          } else if (res.data.status === 401) {
            toast.error("Sorry, your session is wrong. Please, relogin.")
          } else if (res.data.status === 404) {
            toast.error('Please, login.')
          }
        })
          .catch((error) => {
            if (error) {
              toast.error("Internal server error. Failed to load suggested articles.")
              console.error('Internal server error:', error)
            }
          })
        
        toast.warn('Can`t be deleted. This article has already been published.')
      }
    })
      .catch((error) => {
        if (error) {
          toast.error("Internal server error.")
          console.error('Internal server error:', error)
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
      if (res.data.body.access_token) {
        localStorage.setItem('access_token', 'Bearer ' + res.data.body.access_token)
      }
      if (res.data.status === 500) {
        toast.error("Internal server error. Please try later.")
      } else if (res.data.status === 401) {
        toast.error("Sorry, your session is wrong. Please, relogin.")
      } else if (res.data.status === 404) {
        toast.error('Please, login.')
      } else if (res.data.status === 206) {
        toast.info("We already know about this article. Thank you!")
      } else if (res.data.status === 204) {
        toast.warn("This article is not suitable, sorry. Please, try another one.")
      } else if (res.data.status === 205) {
        toast.info('You have successfully changed an article!')
        this.setState({ articles: [] })
      } else if (res.data.status === 202) {
        toast.info('You have successfully changed an article!')
        this.setState({ articles: res.data.body.articles })
      } else if (res.data.status === 403) {
        axios.get(baseurl, config).then((res) => {
          if (res.data.body.access_token) {
            localStorage.setItem('access_token', 'Bearer ' + res.data.body.access_token)
          }
          
          if (res.data.status === 200) {
            this.setState({ articles: res.data.body.articles })
          } else if (res.data.status === 204) {
            this.setState({ articles: [] })
          } else if (res.data.status === 401) {
            toast.error("Sorry, your session is wrong. Please, relogin.")
          } else if (res.data.status === 404) {
            toast.error('Please, login.')
          }
        })
          .catch((error) => {
            if (error) {
              toast.error("Internal server error. Failed to load suggested articles.")
              console.error('Internal server error:', error)
            }
          })

        toast.warn('Can`t be changed. This article has already been published.')
      }
    })
      .catch((error) => {
        if (error) {
          toast.error("Internal server error. Failed to load suggested articles.")
          console.error('Internal server error:', error)
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
      if (res.data.body.access_token) {
        localStorage.setItem('access_token', 'Bearer ' + res.data.body.access_token)
      }
      if (res.data.status === 500) {
        toast.error("Internal server error. Please try later.")
      } else if (res.data.status === 401) {
        toast.error("Sorry, your session is wrong. Please, relogin.")
      } else if (res.data.status === 404) {
        toast.error('Please, login.')
      } else if (res.data.status === 206) {
        toast.info("We already know about this article. Thank you!")
      } else if (res.data.status === 204) {
        toast.warn("This article is not suitable, sorry. Please, try another one.")
      } else if (res.data.status === 205) {
        toast.success('You have successfully suggested an article!')
        this.setState({ articles: [] })
      } else if (res.data.status === 201) {
        toast.success('You have successfully suggested an article!')
        this.setState({ articles: res.data.body.articles })
      }
    })
      .catch((error) => {
        if (error) {
          toast.error("Internal server error. Failed to load suggested articles.")
          console.error('Internal server error:', error)
        }
      })
  }
}