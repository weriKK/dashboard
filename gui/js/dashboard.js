/*
dragula([
	document.getElementById('1'),
	document.getElementById('2'),
	document.getElementById('3')
])

.on('drag', function(el) {

	// add 'is-moving' class to element being dragged
	el.classList.add('is-moving');
})
.on('dragend', function(el) {

	// remove 'is-moving' class from element after dragging has stopped
	el.classList.remove('is-moving');

	// add the 'is-moved' class for 600ms then remove it
	window.setTimeout(function() {
		el.classList.add('is-moved');
		window.setTimeout(function() {
			el.classList.remove('is-moved');
		}, 200);
	}, 100);
});
*/




/*
<li class="column column-red">
	<ul class="drag-item-list" id="1">
		<li class="drag-item">
			<span class="drag-item-header"><h2>Test 1</h2></span>
			<ul>
				<li><a href="">AA AA AA AA AA</a></li>
				<li><a href="">BA BA BA BA BA</a></li>
				<li><a href="">CA CA CA CA CA</a></li>
				<li><a href="">DA DA DA DA DA</a></li>
			</ul>
		</li>
	</ul>
</li>
*/

'use struct';

function Column(props) {
	return (
    <li className={"column column-" + props.color}>
			<FeedBoxList id={props.id} feeds={props.feeds} />
		</li>
  );
}

function FeedBoxList(props) {
	const feedBoxes = props.feeds.map((feed) =>
  	<FeedBox feed={feed} key={feed.title} />
	);

	return (
	  <ul className="drag-item-list" id={props.id}>
	    {feedBoxes}
	  </ul>
  );
}

function FeedBox(props) {
  const feedItems = props.feed.items.map((item) =>
    <FeedItem url={item.url} title={item.title} key={item.title} />
  );
  return (
    <li className="drag-item">
      <FeedBoxHeader text={props.feed.title} />
      <ul>
        {feedItems}
      </ul>
    </li>
  );
}

function FeedBoxHeader(props) {
  return <span className="drag-item-header"><h2>{props.text}</h2></span>;
}

function FeedItem(props) {
  return <li><a href={props.url} target="_blank">{props.title}</a></li>;
}

const data = [
  {
    color: "red",
    feeds: [
      {title: "Title 1", url: "url1", items: [
        {url: "itemurl1", title: "First link"},
        {url: "itemurl2", title: "Second link"},
        {url: "itemurl3", title: "Third link"},
      ]},
      {title: "Title 2", url: "url2", items: [
        {url: "itemurl1", title: "First link"},
        {url: "itemurl2", title: "Second link"},
        {url: "itemurl3", title: "Third link"},
        {url: "itemurl4", title: "Fourth link"}
      ]},
      {title: "Title 3", url: "url3", items: [
        {url: "itemurl1", title: "First link"},
        {url: "itemurl2", title: "Second link"},
      ]}
    ]
  },
  {
    color: "blue",
    feeds: [
      {title: "Title 4", url: "url4", items: [
        {url: "itemurl1", title: "First link"},
        {url: "itemurl2", title: "Second link"},
        {url: "itemurl3", title: "Third link"},
        {url: "itemurl4", title: "Fourth link"}
      ]},
      {title: "Title 5", url: "url5", items: [
        {url: "itemurl1", title: "First link"},
      ]},
      {title: "Title 6", url: "url6", items: [
        {url: "itemurl1", title: "First link"},
        {url: "itemurl2", title: "Second link"},
        {url: "itemurl3", title: "Third link"},
        {url: "itemurl4", title: "Fourth link"}
      ]},
      {title: "Title 7", url: "url7", items: [
        {url: "itemurl1", title: "First link"},
        {url: "itemurl3", title: "Third link"},
        {url: "itemurl4", title: "Fourth link"}
      ]}
    ]
  }
];

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {data: props.data};
  }

  componentDidMount() {
    // runs after the component output has been rendered to the DOM
		// var container = ReactDOM.findDOMNode(this);
		// reactDragula([container]);

		var feedLists = document.querySelectorAll('.drag-item-list');
		reactDragula(Array.from(feedLists))
		.on('drag', function(el) { el.classList.add('is-moving'); })
		.on('dragend', function(el) {
			el.classList.remove('is-moving');

			// add the 'is-moved' class for 600ms then remove it
			window.setTimeout(function() {
				el.classList.add('is-moved');
				window.setTimeout(function() {
					el.classList.remove('is-moved');
				}, 200);
			}, 100);
		});
  }

  componentWillUnmount() {
    // runs before the component is removed from the DOM
  }

  render() {
    const columns = this.state.data.map((column, index) =>
      <Column id={index} color={column.color} feeds={column.feeds} key={index}/>
    );

    return (
      <ul id="columns">
        {columns}
      </ul>
    );
  }
}

ReactDOM.render(<App data={data}/>, document.getElementById('root'));





/*
'use struct';

const e = React.createElement;

class AppComponent extends React.Component {
	constructor(props) {
		super(props);
	}

	render() {
		return e(
			"div",
			{className: "container"},
			React.createElement("div", null, "Swap me around"),
			React.createElement("div", null, "Swap her around"),
			React.createElement("div", null, "Swap him around"),
			React.createElement("div", null, "Swap us around"),
		);
	}

	componentDidMount() {
		var container = ReactDOM.findDOMNode(this);
		reactDragula([container]);
	}
}

const domContainer = document.querySelector('#dashboard');
ReactDOM.render(e(AppComponent), domContainer);

//React.render(React.createElement(App, null), document.getElementById('dashboard'));
*/


// 'use strict';

// const e = React.createElement;

// class LikeButton extends React.Component {
//   constructor(props) {
//     super(props);
//     this.state = { liked: false };
//   }

//   render() {
//     if (this.state.liked) {
//       return 'You liked this.';
//     }

//     return e(
//       'button',
//       { onClick: () => this.setState({ liked: true }) },
//       'Like'
//     );
//   }
// }

// const domContainer = document.querySelector('#dashboard');
// console.log(domContainer);
// ReactDOM.render(e(LikeButton), domContainer);
